package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func getAllMetricsHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/html") //iter8 fix

	io.WriteString(resp, "<h2>All metrics</h2>")

	io.WriteString(resp, "<h2>Counter</h2>")
	counterAll, _ := MetricStorage.CounterAll()
	for name, val := range counterAll {
		io.WriteString(resp, fmt.Sprintf("<div>%s: %d</div>", name, val))
	}

	io.WriteString(resp, "<h2>Gauge</h2>")
	gaugeAll, _ := MetricStorage.GaugeAll()
	for name, val := range gaugeAll {
		io.WriteString(resp, fmt.Sprintf("<div>%s: %f</div>", name, val))
	}
}

func getValueHandler(resp http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	if name == "" {
		http.Error(resp, "Bad name", http.StatusNotFound)
		return
	}

	switch chi.URLParam(req, "type") {
	case "counter":
		val, err := MetricStorage.CounterGet(name)
		if err != nil {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		io.WriteString(resp, fmt.Sprintf("%d", val))

	case "gauge":
		val, err := MetricStorage.GaugeGet(name)
		if err != nil {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		io.WriteString(resp, fmt.Sprintf("%g", val))

	default:
		http.Error(resp, "invalid metric type", http.StatusNotFound)
		return
	}
}

func postUpdateHandler(resp http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	switch strings.ToLower(chi.URLParam(req, "type")) {
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(resp, "invalid counter value", http.StatusBadRequest)
			return
		}
		MetricStorage.CounterInc(name, v)

	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(resp, "invalid gauge value", http.StatusBadRequest)
			return
		}
		MetricStorage.GaugeSet(name, v)

	default:
		http.Error(resp, "invalid metric type", http.StatusBadRequest)
		return
	}

	resp.WriteHeader(http.StatusOK)
	io.WriteString(resp, "OK")
}

func postValueJSONHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metric, err := decodeMetricsRequest(req)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		Log.Warn(err.Error())
		return
	}

	switch metric.MType {
	case "counter":
		val, _ := MetricStorage.CounterGet(metric.ID)
		metric.Delta = &val

	case "gauge":
		val, _ := MetricStorage.GaugeGet(metric.ID)
		metric.Value = &val

	default:
		httpErrorJSON(resp, "invalid metric type", http.StatusNotFound)
		return
	}

	b, err := json.Marshal(metric)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusOK)
	_, err = resp.Write(b)
	if err != nil {
		Log.Error(err.Error())
	}
}

func postUpdateJSONHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metric, err := decodeMetricsRequest(req)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		Log.Warn(err.Error())
		return
	}

	if err := persistMetric(metric); err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		return
	}

	b, err := json.Marshal(resp)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusOK)
	_, err = resp.Write(b)
	if err != nil {
		Log.Error(err.Error())
	}
}

func postUpdateBatchJSONHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metricList, err := decodeMetricsBatchRequest(req)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		Log.Warn(err.Error())
		return
	}

	for _, m := range metricList {
		if err := persistMetric(&m); err != nil {
			httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
			return
		}
	}

	b, err := json.Marshal(resp)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.WriteHeader(http.StatusOK)
	_, err = resp.Write(b)
	if err != nil {
		Log.Error(err.Error())
	}
}

func getPingHandler(resp http.ResponseWriter, req *http.Request) {
	ctx, cancel := context.WithTimeout(req.Context(), 2*time.Second)
	defer cancel()

	err := Database.PingContext(ctx)
	if err != nil {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "text/html")
	resp.WriteHeader(http.StatusOK)
	resp.Write([]byte("OK"))
}

func decodeMetricsRequest(req *http.Request) (*common.Metrics, error) {
	if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
		return nil, errors.New("invalid Content-Type header")
	}

	var metric = new(common.Metrics)
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&metric); err != nil {
		return nil, errors.New("failed to decode request body")
	}
	defer req.Body.Close()

	if metric.ID == "" || metric.MType == "" {
		return nil, errors.New("mandatory properties not specified")
	}

	return metric, nil
}

func decodeMetricsBatchRequest(req *http.Request) ([]common.Metrics, error) {
	if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
		return nil, errors.New("invalid Content-Type header")
	}

	metricList := make([]common.Metrics, 0, 100)
	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&metricList); err != nil {
		return nil, errors.New("failed to decode request body")
	}
	defer req.Body.Close()

	return metricList, nil
}

func httpErrorJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

// persistMetric Сохраняем метрику в хранилище.
func persistMetric(m *common.Metrics) error {
	switch strings.ToLower(m.MType) {
	case "counter":
		if *m.Delta < 0 {
			return errors.New("counter delta must be positive")
		}
		MetricStorage.CounterInc(m.ID, *m.Delta)

	case "gauge":
		MetricStorage.GaugeSet(m.ID, *m.Value)

	default:
		return errors.New("invalid m type")
	}

	return nil
}
