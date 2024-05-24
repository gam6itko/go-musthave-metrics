package main

import (
	"context"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	sync2 "github.com/gam6itko/go-musthave-metrics/internal/sync"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"sync"
)

// todo функционал очень похожий
func initGRPCClient(ctx context.Context, wg *sync.WaitGroup) chan<- []*common.Metrics {
	ch := make(chan []*common.Metrics)

	// устанавливаем соединение с сервером
	conn, err := grpc.Dial(
		AppConfig.Address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)

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
