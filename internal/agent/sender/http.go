package sender

import (
	"bytes"
	"crypto/hmac"
	cryrand "crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	"io"
	"log"
	"net/http"
)

type HTTPSender struct {
	client    *http.Client
	address   string
	publicKey *rsa.PublicKey
}

func NewHTTPSender(client *http.Client) *HTTPSender {
	return &HTTPSender{client: client}
}

func (ths HTTPSender) Send(metricList []*common.Metrics) error {
	requestBody := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(requestBody)
	if err := encoder.Encode(metricList); err != nil {
		return err
	}

	//todo middleware
	if ths.publicKey != nil {
		hash := sha256.New()
		enc, err := rsautils.EncryptOAEP(hash, cryrand.Reader, ths.publicKey, requestBody.Bytes(), nil)
		if err != nil {
			log.Fatal(err)
		}
		requestBody = bytes.NewBuffer(enc)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s/updates/", ths.address),
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
