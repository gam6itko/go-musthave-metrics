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
	for name, val := range memory.CounterAll() {
		io.WriteString(resp, fmt.Sprintf("%s: %d\n", name, val))
	}
	for name, val := range memory.GaugeAll() {
		io.WriteString(resp, fmt.Sprintf("%s: %f\n", name, val))
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
	metric, err := decodeMetricsRequest(req)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		Log.Warn(err.Error())
		return
	}

	switch metric.MType {
	case "counter":
		val, exists := memory.CounterGet(metric.ID)
		if !exists {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		metric.Delta = val

	case "gauge":
		val, exists := memory.GaugeGet(metric.ID)
		if !exists {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		metric.Value = val

	default:
		http.Error(resp, "invalid metric type", http.StatusNotFound)
		return
	}

	encoder := json.NewEncoder(resp)
	if err := encoder.Encode(metric); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
}

func postUpdateJSONHandler(resp http.ResponseWriter, req *http.Request) {
	metric, err := decodeMetricsRequest(req)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusBadRequest)
		Log.Warn(err.Error())
		return
	}

	switch strings.ToLower(metric.MType) {
	case "counter":
		if metric.Delta <= 0 {
			http.Error(resp, "counter delta must be positive", http.StatusBadRequest)
			return
		}
		memory.CounterInc(metric.ID, metric.Delta)

	case "gauge":
		memory.GaugeSet(metric.ID, metric.Value)

	default:
		http.Error(resp, "invalid metric type", http.StatusBadRequest)
		return
	}

	encoder := json.NewEncoder(resp)
	if err := encoder.Encode(metric); err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
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
