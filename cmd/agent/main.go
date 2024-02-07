package main

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	commonFlags "github.com/gam6itko/go-musthave-metrics/internal/common/flags"
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
}

var _serverAddr commonFlags.NetAddress
var _stat metrics
var _reportInterval uint
var _pollInterval uint
var _key string

func init() {
	_serverAddr = commonFlags.NewNetAddr("localhost", 8080)

	_ = flag.Value(&_serverAddr)
	flag.Var(&_serverAddr, "a", "Server address host:port")
	reportIntervalF := flag.Uint("r", 10, "Report interval")
	pollIntervalF := flag.Uint("p", 2, "Poll interval")
	keyF := flag.String("k", "", "Encryption key")
	flag.Parse()

	_reportInterval = *reportIntervalF
	_pollInterval = *pollIntervalF
	_key = *keyF

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

	go func() {
		defer wg.Done()

		for {
			func() {
				mux.Lock()
				defer mux.Unlock()

				runtime.ReadMemStats(&_stat.MemStats)
				_stat.PollCount++
				_stat.RandomValue = rand.Float64()
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

				metricList := make([]common.Metrics, 0, len(GaugeToSend)+2)

				for _, gName := range GaugeToSend {
					f := reflect.Indirect(refValue).FieldByName(gName)

					m := common.Metrics{
						ID:    gName,
						MType: "gauge",
					}
					if f.CanInt() {
						m.ValueForRef = float64(f.Int())
					} else if f.CanUint() {
						m.ValueForRef = float64(f.Uint())
					} else if f.CanFloat() {
						m.ValueForRef = f.Float()
					} else {
						log.Printf("failed to get gauge value `%s`", gName)
						continue
					}

					m.Value = &m.ValueForRef

					metricList = append(metricList, m)
				}

				m := common.Metrics{
					ID:          "PollCount",
					MType:       "counter",
					DeltaForRef: _stat.PollCount,
				}
				m.Delta = &m.DeltaForRef
				metricList = append(metricList, m)

				if err := sendMetrics(&httpClient, &metricList); err != nil {
					log.Printf("errors making http request: %s\n", err)
				}
			}()
		}
	}()
}

func sendMetrics(httpClient *http.Client, metricList *[]common.Metrics) error {
	requestBody := bytes.NewBuffer([]byte{})
	encoder := json.NewEncoder(requestBody)
	if err := encoder.Encode(metricList); err != nil {
		return err
	}

	// request send
	req, err := http.NewRequest(
		http.MethodPost,
		fmt.Sprintf("http://%s/updates/", _serverAddr.String()),
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
}
