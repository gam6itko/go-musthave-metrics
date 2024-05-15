package main

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/agent/config"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	sync2 "github.com/gam6itko/go-musthave-metrics/internal/sync"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"strconv"
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

var _cfg config.Config
var _stat metrics
var _rsaPublicKey *rsa.PublicKey

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	_cfg = initConfig()

	if _cfg.RSAPublicKey != "" {
		_rsaPublicKey = loadPublicKey(_cfg.RSAPublicKey)
	}

	_stat = metrics{
		PollCount: 0,
	}
}

// loadPublicKey загружает publicKey из файла.
func loadPublicKey(path string) *rsa.PublicKey {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return rsautils.BytesToPublicKey(b)
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

					runtime.ReadMemStats(&_stat.MemStats)
					// custom
					_stat.PollCount++
					_stat.RandomValue = rand.Float64()
				}()
				time.Sleep(time.Duration(_cfg.PollInterval) * time.Second)
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
					_stat.TotalMemory = v.Total
					_stat.FreeMemory = v.Free
					util, err := cpu.Percent(time.Second, true)
					if err != nil {
						log.Printf("cpu error: %s", err)
					}
					_stat.CPUUtilization = util
				}()
				time.Sleep(time.Duration(_cfg.PollInterval) * time.Second)
			}
		}(ctx, &wg)

		wg.Add(1)
		go startReporting(ctx, &mux, &wg)

		wg.Wait()
		log.Printf("DEBUG. exit from server")
		if err2 := server.Shutdown(context.Background()); err2 != nil {
			log.Printf("ERROR. server shutdown error: %s", err2)
		}
	}()

	if err := server.ListenAndServe(); err != nil {
		log.Printf("http server returns error: %s", err)
	}
}

// startReporting запустить сбор и отправку метрик.
func startReporting(ctx context.Context, mux *sync.RWMutex, wg *sync.WaitGroup) {
	defer wg.Done()

	GaugeToSend := []string{
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

	httpClient := http.Client{
		Timeout: 30 * time.Second,
	}

	sleepDuration := time.Duration(_cfg.ReportInterval) * time.Second
infLoop:
	for {
		select {
		default:
		case <-ctx.Done():
			log.Printf("DEBUG. exit from go reporting loop")
			break infLoop
		}

		time.Sleep(sleepDuration)
		log.Printf("sending metrics: %d\n", _stat.PollCount)

		func() {
			mux.RLock()
			defer mux.RUnlock()

			refValue := reflect.ValueOf(_stat)

			metricList := make([]*common.Metrics, 0, len(GaugeToSend)+2)

			for _, gName := range GaugeToSend {
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
					Delta: common.Int64Ref(_stat.PollCount),
				},
			)

			//gopsutils cpu
			for i, val := range _stat.CPUUtilization {
				metricList = append(
					metricList,
					&common.Metrics{
						ID:    fmt.Sprintf("CPUutilization%d", i),
						MType: "gauge",
						Value: common.Float64Ref(val),
					},
				)
			}

			if err := sendMetrics(&httpClient, metricList); err != nil {
				log.Printf("errors making http request: %s\n", err)
			}
		}()
	}
}

// sendMetrics отправляет мертрики на сервер.
func sendMetrics(httpClient *http.Client, metricList []*common.Metrics) error {
	g := new(errgroup.Group)

	var semaphore sync2.ISemaphore
	semaphore = &sync2.NullSemaphore{}
	if _cfg.RateLimit > 0 {
		semaphore = sync2.NewSemaphore(_cfg.RateLimit)
	}

	for _, m := range metricList {
		oneMetric := m
		g.Go(func() error {
			semaphore.Acquire()
			defer semaphore.Release()

			requestBody := bytes.NewBuffer([]byte{})
			encoder := json.NewEncoder(requestBody)
			if err := encoder.Encode(metricList); err != nil {
				return err
			}

			// request send
			var valueStr string
			switch oneMetric.MType {
			case string(common.Counter):
				valueStr = strconv.FormatInt(*oneMetric.Delta, 10)
			case string(common.Gauge):
				valueStr = strconv.FormatFloat(*oneMetric.Value, 'f', 10, 64)
			default:
				return errors.New("invalid MType")
			}

			if _rsaPublicKey != nil {
				hash := sha512.New()
				enc, err := rsautils.EncryptOAEP(hash, requestBody, _rsaPublicKey, requestBody.Bytes(), nil)
				if err != nil {
					log.Fatal(err)
				}
				requestBody = bytes.NewBuffer(enc)
			}

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf(
					"http://%s/update/%s/%s/%s",
					_cfg.Address,
					oneMetric.MType,
					oneMetric.ID,
					valueStr,
				),
				requestBody,
			)
			if err != nil {
				return err
			}

			req.Header.Set("Content-Type", "application/json")

			if _cfg.SignKey != "" {
				// подписываем алгоритмом HMAC, используя SHA-256
				h := hmac.New(sha256.New, []byte(_cfg.SignKey))
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

			resp.Body.Close()

			return nil
		})
	}

	return g.Wait()
}
