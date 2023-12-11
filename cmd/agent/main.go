package main

import (
	"flag"
	"fmt"
	commonFlags "github.com/gam6itko/go-musthave-metrics/internal/common/flags"
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
}

var serverAddr commonFlags.NetAddress
var stat metrics
var reportInterval uint
var pollInterval uint

func init() {
	serverAddr = commonFlags.NewNetAddr("localhost", 8080)

	_ = flag.Value(&serverAddr)
	flag.Var(&serverAddr, "a", "Server address host:port")
	reportInterval = *flag.Uint("r", 10, "Report interval")
	pollInterval = *flag.Uint("p", 2, "Poll interval")
	flag.Parse()

	// read from env
	if envVal := os.Getenv("ADDRESS"); envVal != "" {
		if err := serverAddr.FromString(envVal); err != nil {
			panic(err)
		}
	}
	if envVal := os.Getenv("REPORT_INTERVAL"); envVal != "" {
		if val, err := strconv.ParseUint(envVal, 10, 32); err == nil {
			reportInterval = uint(val)
		}
	}
	if envVal := os.Getenv("POLL_INTERVAL"); envVal != "" {
		if val, err := strconv.ParseUint(envVal, 10, 32); err == nil {
			pollInterval = uint(val)
		}
	}

	stat = metrics{
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

	go func() {
		defer wg.Done()

		for {
			func() {
				mux.Lock()
				defer mux.Unlock()

				runtime.ReadMemStats(&stat.MemStats)
				stat.PollCount++
				stat.RandomValue = rand.Float64()
			}()
			time.Sleep(time.Duration(pollInterval) * time.Second)
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
	}

	go func() {
		defer wg.Done()

		httpClient := http.Client{
			Timeout: 30 * time.Second,
		}

		for {
			time.Sleep(time.Duration(reportInterval) * time.Second)
			fmt.Printf("sending metrics: %d\n", stat.PollCount)

			func() {
				mux.RLock()
				defer mux.RUnlock()

				refValue := reflect.ValueOf(stat)

				for _, gName := range GaugeToSend {
					f := reflect.Indirect(refValue).FieldByName(gName)

					var valueStr string
					if f.CanInt() {
						valueStr = fmt.Sprintf("%d", f.Int())
					} else if f.CanUint() {
						valueStr = fmt.Sprintf("%d", f.Uint())
					} else if f.CanFloat() {
						valueStr = fmt.Sprintf("%f", f.Float())
					} else {
						fmt.Printf("failed to get gauge value `%s`", gName)
						continue
					}

					req, err := http.NewRequest(
						http.MethodPost,
						fmt.Sprintf("http://%s/update/gauge/%s/%s", serverAddr.String(), gName, valueStr),
						nil,
					)
					if err != nil {
						fmt.Printf("client: errors build http request: %s\n", err)
					}

					req.Header.Set("Content-Type", "text/plain")

					resp, err := httpClient.Do(req)
					if err != nil {
						fmt.Printf("client: errors making http request: %s\n", err)
						break
					}
					defer resp.Body.Close()
				}

				// PollCount
				req, err := http.NewRequest(
					http.MethodPost,
					fmt.Sprintf("http://%s/update/counter/%s/%d", serverAddr.String(), "PollCount", stat.PollCount),
					nil,
				)
				if err != nil {
					fmt.Printf("client: errors build http request: %s\n", err)
				}

				req.Header.Set("Content-Type", "text/plain")

				resp, err := httpClient.Do(req)
				if err != nil {
					fmt.Printf("client: errors making http request: %s\n", err)
				}
				defer resp.Body.Close()
			}()
		}
	}()
}
