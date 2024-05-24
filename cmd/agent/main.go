package main

import (
	"context"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/agent/config"
	http2 "github.com/gam6itko/go-musthave-metrics/internal/agent/http"
	"github.com/gam6itko/go-musthave-metrics/internal/agent/sender"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	sync2 "github.com/gam6itko/go-musthave-metrics/internal/sync"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"syscall"
	"time"
)

type metrics struct {
	// Second
	CPUUtilization []float64

	PollCount   int64
	RandomValue float64
	//gopsutil
	TotalMemory uint64
	FreeMemory  uint64

	runtime.MemStats
}

var AppConfig config.Config
var Stat metrics

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	AppConfig = initConfig()

	Stat = metrics{
		PollCount: 0,
	}
}

func main() {
	server := &http.Server{
		Addr:    ":8081",
		Handler: http.DefaultServeMux,
	}

	go func() {
		ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

		mux := sync.RWMutex{}

		wg := sync.WaitGroup{}

		// runtime
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				select {
				default:
				case <-ctx.Done():
					log.Printf("DEBUG. exit from go runtime:")
					return // exit from goroutine
				}

				func() {
					mux.Lock()
					defer mux.Unlock()

					runtime.ReadMemStats(&Stat.MemStats)
					// custom
					Stat.PollCount++
					Stat.RandomValue = rand.Float64()
				}()
				time.Sleep(time.Duration(AppConfig.PollInterval) * time.Second)
			}
		}(ctx, &wg)

		// gopsutil
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					log.Printf("DEBUG. exit from go gopsutil:")
					return //exit from goroutine
				default:
					// go further
				}
				func() {
					mux.Lock()
					defer mux.Unlock()

					v, _ := mem.VirtualMemory()
					Stat.TotalMemory = v.Total
					Stat.FreeMemory = v.Free
					util, err := cpu.Percent(time.Second, true)
					if err != nil {
						log.Printf("cpu error: %s", err)
					}
					Stat.CPUUtilization = util
				}()
				time.Sleep(time.Duration(AppConfig.PollInterval) * time.Second)
			}
		}(ctx, &wg)

		wg.Add(1)
		ch := initSending(ctx, &wg)
		go startReporting(ctx, &mux, &wg, ch)

		wg.Wait()

		log.Printf("DEBUG. stop http server")
		if err2 := server.Shutdown(context.Background()); err2 != nil {
			log.Printf("ERROR. server shutdown error: %s", err2)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Printf("http server returns error: %s", err)
	}
}

// initSending создаёт и прослушивает канал для отправки метрик.
func initSending(ctx context.Context, wg *sync.WaitGroup) chan<- []*common.Metrics {
	ch := make(chan []*common.Metrics)

	var s sender.ISender
	if AppConfig.UseGRPC {
		s = sender.NewGRPCSender(AppConfig.Address)
	} else {
		httpClient := buildHTTPClient()
		s = sender.NewHTTPSender(httpClient, AppConfig.Address)
	}

	wg.Add(1)
	go func(s sender.ISender) {
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
				log.Printf("DEBUG. exit from HTTP s")
				return // exit from goroutine
			}

			go func(metricList []*common.Metrics) {
				semaphore.Acquire()
				defer semaphore.Release()

				if err := s.Send(metricList); err != nil {
					log.Printf("Failed to send metrics: %v", err)
				}
			}(metricList)
		}
	}(s)

	return ch
}

func buildHTTPClient() http2.IClient {
	var client http2.IClient
	client = &http.Client{
		Timeout: 30 * time.Second,
	}

	if AppConfig.RSAPublicKey != "" {
		b, err := os.ReadFile(AppConfig.RSAPublicKey)
		if err != nil {
			log.Fatal(err)
		}
		client = http2.NewEncryptDecorator(client, rsautils.BytesToPublicKey(b))
	}

	if AppConfig.XRealIP != "" {
		ip := net.ParseIP(AppConfig.XRealIP)
		if ip == nil {
			log.Fatalf("Invalid XRealIP string %s", AppConfig.XRealIP)
		}
		client = http2.NewXRealIPDecorator(client, ip)
	}

	if AppConfig.SignKey != "" {
		client = http2.NewSignDecorator(client, AppConfig.SignKey)
	}

	return client
}

// startReporting запустить сбор и отправку метрик.
func startReporting(ctx context.Context, mux *sync.RWMutex, wg *sync.WaitGroup, ch chan<- []*common.Metrics) {
	defer wg.Done()

	gaugeToSend := []string{
		"Alloc",
		"BuckHashSys",
		"Frees",
		"GCCPUFraction",
		"GCSys",
		"HeapAlloc",
		"HeapIdle",
		"HeapInuse",
		"HeapObjects",
		"HeapReleased",
		"HeapSys",
		"LastGC",
		"Lookups",
		"MCacheInuse",
		"MCacheSys",
		"MSpanInuse",
		"MSpanSys",
		"Mallocs",
		"NextGC",
		"NumForcedGC",
		"NumGC",
		"OtherSys",
		"PauseTotalNs",
		"StackInuse",
		"StackSys",
		"Sys",
		"TotalAlloc",
		// custom
		"RandomValue",
		//gopsutils
		"TotalMemory",
		"FreeMemory",
	}

	sleepDuration := time.Duration(AppConfig.ReportInterval) * time.Second
infLoop:
	for {
		select {
		default:
		case <-ctx.Done():
			log.Printf("DEBUG. exit from go reporting loop")
			break infLoop
		}

		time.Sleep(sleepDuration)
		log.Printf("sending metrics: %d\n", Stat.PollCount)

		ch <- func() []*common.Metrics {
			mux.RLock()
			defer mux.RUnlock()

			refValue := reflect.ValueOf(Stat)

			metricList := make([]*common.Metrics, 0, len(gaugeToSend)+2)

			for _, gName := range gaugeToSend {
				f := reflect.Indirect(refValue).FieldByName(gName)

				m := &common.Metrics{
					ID:    gName,
					MType: "gauge",
				}
				if f.CanInt() {
					m.Value = common.Float64Ref(float64(f.Int()))
				} else if f.CanUint() {
					m.Value = common.Float64Ref(float64(f.Uint()))
				} else if f.CanFloat() {
					m.Value = common.Float64Ref(f.Float())
				} else {
					log.Printf("failed to get gauge value `%s`", gName)
					continue
				}

				metricList = append(metricList, m)
			}

			metricList = append(
				metricList,
				&common.Metrics{
					ID:    "PollCount",
					MType: "counter",
					Delta: common.Int64Ref(Stat.PollCount),
				},
			)

			//gopsutils cpu
			for i, val := range Stat.CPUUtilization {
				metricList = append(
					metricList,
					&common.Metrics{
						ID:    fmt.Sprintf("CPUutilization%d", i),
						MType: "gauge",
						Value: common.Float64Ref(val),
					},
				)
			}

			return metricList
		}()
	}
}
