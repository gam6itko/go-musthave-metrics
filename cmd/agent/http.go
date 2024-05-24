package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	cryrand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	sync2 "github.com/gam6itko/go-musthave-metrics/internal/sync"
	"io"
	"log"
	"net/http"
	"sync"
	"time"
)

// initHTTPSender возвращает канал в который должны быть записаны метрики для отправки на HTTP-сервер.
func initHTTPSender(ctx context.Context, wg *sync.WaitGroup) chan<- []*common.Metrics {
	ch := make(chan []*common.Metrics)

	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()

		var semaphore sync2.ISemaphore
		semaphore = &sync2.NullSemaphore{}
		if AppConfig.RateLimit > 0 {
			semaphore = sync2.NewSemaphore(AppConfig.RateLimit)
		}

		for metricList := range ch {
			select {
			default:
			case <-ctx.Done():
				log.Printf("DEBUG. exit from HTTP sender")
				return // exit from goroutine
			}

			go func(metricList []*common.Metrics) {
				semaphore.Acquire()
				defer semaphore.Release()

				if err := sendHTTP(&httpClient, metricList); err != nil {
					log.Printf("Failed to send metrics: %v", err)
				}
			}(metricList)
		}
	}()

	return ch
}

func sendHTTP(httpClient *http.Client, metricList []*common.Metrics) error {
	requestBody := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(requestBody)
	if err := encoder.Encode(metricList); err != nil {
		return err
	}

	//todo middleware
	if RSAPublicKey != nil {
		hash := sha256.New()
		enc, err := rsautils.EncryptOAEP(hash, cryrand.Reader, RSAPublicKey, requestBody.Bytes(), nil)
		if err != nil {
			log.Fatal(err)
		}
		requestBody = bytes.NewBuffer(enc)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s/updates/", AppConfig.Address),
		requestBody,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if AppConfig.XRealIP != "" {
		req.Header.Set("X-Real-IP", AppConfig.XRealIP)
	}

	if AppConfig.SignKey != "" {
		// подписываем алгоритмом HMAC, используя SHA-256
		h := hmac.New(sha256.New, []byte(AppConfig.SignKey))
		if _, wErr := h.Write(requestBody.Bytes()); wErr != nil {
			return wErr
		}
		dst := h.Sum(nil)

		base64Enc := base64.StdEncoding.EncodeToString(dst)
		req.Header.Set("HashSHA256", base64Enc)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}

	defer func() {
		if err = resp.Body.Close(); err != nil {
			log.Printf("ERROR. close body: %s", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		bMsg, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		log.Printf("WARNING. status is not 200. Body: %s", string(bMsg))
	}

	return nil
}
