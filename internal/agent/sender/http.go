package sender

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	http2 "github.com/gam6itko/go-musthave-metrics/internal/agent/http"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"io"
	"log"
	"net/http"
)

type HTTPSender struct {
	httpClient http2.IClient
	address    string
}

func NewHTTPSender(client http2.IClient, address string) *HTTPSender {
	return &HTTPSender{
		httpClient: client,
		address:    address,
	}
}

func (ths HTTPSender) Send(ctx context.Context, metricList []*common.Metrics) error {
	requestBody := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(requestBody)
	if err := encoder.Encode(metricList); err != nil {
		return err
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

	resp, err := ths.httpClient.Do(req.WithContext(ctx))
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
