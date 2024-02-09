package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	commonFlags "github.com/gam6itko/go-musthave-metrics/internal/common/flags"
	sync2 "github.com/gam6itko/go-musthave-metrics/internal/sync"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sync/errgroup"
	"log"
	"math/rand"
	"net/http"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"sync"
	"time"
)

type metrics struct {
	runtime.MemStats
	PollCount   int64
	RandomValue float64
	//gopsutil
	TotalMemory uint64
	FreeMemory  uint64
	// Second
	CPUUtilization []float64
}

var _serverAddr commonFlags.NetAddress
var _stat metrics
var _reportInterval uint
var _pollInterval uint
var _key string
var _rateLimit uint

func init() {
	_serverAddr = commonFlags.NewNetAddr("localhost", 8080)

	_ = flag.Value(&_serverAddr)
	flag.Var(&_serverAddr, "a", "Server address host:port")
	reportIntervalF := flag.Uint("r", 10, "Report interval")
	pollIntervalF := flag.Uint("p", 2, "Poll interval")
	keyF := flag.String("k", "", "Encryption key")
	rateLimitF := flag.Uint("l", 0, "Request rate limit")
	flag.Parse()

	_reportInterval = *reportIntervalF
	_pollInterval = *pollIntervalF
	_key = *keyF
	_rateLimit = *rateLimitF

	// read from env
	if envVal := os.Getenv("ADDRESS"); envVal != "" {
		if err := _serverAddr.FromString(envVal); err != nil {
			panic(err)
		}
	}
	if envVal := os.Getenv("REPORT_INTERVAL"); envVal != "" {
		if val, err := strconv.ParseUint(envVal, 10, 32); err == nil {
			_reportInterval = uint(val)
		}
	}
	if envVal := os.Getenv("POLL_INTERVAL"); envVal != "" {
		if val, err := strconv.ParseUint(envVal, 10, 32); err == nil {
			_pollInterval = uint(val)
		}
	}
	if envVal := os.Getenv("KEY"); envVal != "" {
		_key = envVal
	}
	if envVal := os.Getenv("RATE_LIMIT"); envVal != "" {
		if val, err := strconv.ParseUint(envVal, 10, 32); err == nil {
			_rateLimit = uint(val)
		}
	}

	_stat = metrics{
		PollCount: 0,
	}
}

func main() {
	mux := sync.RWMutex{}

	var wg sync.WaitGroup

	startPolling(&wg, &mux)
	startReporting(&wg, &mux)
	wg.Wait()
}

func startPolling(wg *sync.WaitGroup, mux *sync.RWMutex) {
	wg.Add(1)

	// runtime
	go func() {
		defer wg.Done()

		for {
			func() {
				mux.Lock()
				defer mux.Unlock()

				runtime.ReadMemStats(&_stat.MemStats)
				// custom
				_stat.PollCount++
				_stat.RandomValue = rand.Float64()
			}()
			time.Sleep(time.Duration(_pollInterval) * time.Second)
		}
	}()

	// gopsutil
	go func() {
		defer wg.Done()

		for {
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
			time.Sleep(time.Duration(_pollInterval) * time.Second)
		}
	}()
}

func startReporting(wg *sync.WaitGroup, mux *sync.RWMutex) {
	wg.Add(1)

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

	go func() {
		defer wg.Done()

		httpClient := http.Client{
			Timeout: 30 * time.Second,
		}

		sleepDuration := time.Duration(_reportInterval) * time.Second
		for {
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
	}()
}

// Для инкремента 15 мы будем отправлять по одной метрике в разных горутино-запросах.
func sendMetrics(httpClient *http.Client, metricList []*common.Metrics) error {
	g := new(errgroup.Group)

	var semaphore sync2.ISemaphore
	semaphore = &sync2.NullSemaphore{}
	if _rateLimit > 0 {
		semaphore = sync2.NewSemaphore(_rateLimit)
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

			req, err := http.NewRequest(
				http.MethodPost,
				fmt.Sprintf(
					"http://%s/update/%s/%s/%s",
					_serverAddr.String(),
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

			if _key != "" {
				// подписываем алгоритмом HMAC, используя SHA-256
				h := hmac.New(sha256.New, []byte(_key))
				if _, err := h.Write(requestBody.Bytes()); err != nil {
					return err
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
