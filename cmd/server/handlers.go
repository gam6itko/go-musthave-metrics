package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"strconv"
	"strings"
)

func getAllMetricsHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/html") //iter8 fix

	io.WriteString(resp, "<h2>All metrics</h2>")

	io.WriteString(resp, "<h2>Counter</h2>")
	for name, val := range memory.CounterAll() {
		io.WriteString(resp, fmt.Sprintf("<div>%s: %d</div>", name, val))
	}

	io.WriteString(resp, "<h2>Gauge</h2>")
	for name, val := range memory.GaugeAll() {
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
		val, exists := memory.CounterGet(name)
		if !exists {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		io.WriteString(resp, fmt.Sprintf("%d", val))

	case "gauge":
		val, exists := memory.GaugeGet(name)
		if !exists {
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
		memory.CounterInc(name, v)

	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(resp, "invalid gauge value", http.StatusBadRequest)
			return
		}
		memory.GaugeSet(name, v)

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
		val, _ := memory.CounterGet(metric.ID)
		metric.Delta = &val

	case "gauge":
		val, _ := memory.GaugeGet(metric.ID)
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

	switch strings.ToLower(metric.MType) {
	case "counter":
		if *metric.Delta <= 0 {
			httpErrorJSON(resp, "counter delta must be positive", http.StatusBadRequest)
			return
		}
		memory.CounterInc(metric.ID, *metric.Delta)

	case "gauge":
		memory.GaugeSet(metric.ID, *metric.Value)

	default:
		httpErrorJSON(resp, "invalid metric type", http.StatusBadRequest)
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

func decodeMetricsRequest(req *http.Request) (*Metrics, error) {
	if contentType := req.Header.Get("Content-Type"); contentType != "application/json" {
		return nil, errors.New("invalid Content-Type header")
	}

	var metric = new(Metrics)
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

func httpErrorJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}
