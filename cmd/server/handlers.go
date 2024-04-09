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
)

// Health godoc
// @Tags Info
// @Summary Получить все накопленные метрики в формате HTML.
// @ID GetAllMetrics
// @Produce text/html
// @Success 200 {string} string "Метрики"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router / [get]
func getAllMetricsHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/html") //iter8 fix

	io.WriteString(resp, "<h2>All metrics</h2>")

	io.WriteString(resp, "<h2>Counter</h2>")
	counterAll, _ := MetricStorage.CounterAll(req.Context())
	for name, val := range counterAll {
		io.WriteString(resp, fmt.Sprintf("<div>%s: %d</div>", name, val))
	}

	io.WriteString(resp, "<h2>Gauge</h2>")
	gaugeAll, _ := MetricStorage.GaugeAll(req.Context())
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
		val, err := MetricStorage.CounterGet(req.Context(), name)
		if err != nil {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		io.WriteString(resp, fmt.Sprintf("%d", val))

	case "gauge":
		val, err := MetricStorage.GaugeGet(req.Context(), name)
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

// Health godoc
// @Tags Store
// @Summary Сохранить одну метрику.
// @ID UpdateOne
// @Produce text/plain
// @Param type path string true "Metric typ [counter, gauge]"
// @Param name path string true "Metric name"
// @Param value path float64 true "Value"
// @Success 200 {string} string "Метрика сохранена"
// @Failure 400 {string} string "Неверный формат данных"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router /update/{type}/{name}/{value} [get]
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
		if err := MetricStorage.CounterInc(req.Context(), name, v); err != nil {
			http.Error(resp, "fail to counter inc", http.StatusInternalServerError)
			return
		}

	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(resp, "invalid gauge value", http.StatusBadRequest)
			return
		}
		if err := MetricStorage.GaugeSet(req.Context(), name, v); err != nil {
			http.Error(resp, "fail to gauge set", http.StatusInternalServerError)
			return
		}

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
		val, _ := MetricStorage.CounterGet(req.Context(), metric.ID)
		metric.Delta = &val

	case "gauge":
		val, _ := MetricStorage.GaugeGet(req.Context(), metric.ID)
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

	if err := persistMetric(req.Context(), metric); err != nil {
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
		if err := persistMetric(req.Context(), &m); err != nil {
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
	err := Database.PingContext(req.Context())
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
func persistMetric(ctx context.Context, m *common.Metrics) error {
	switch strings.ToLower(m.MType) {
	case "counter":
		if *m.Delta < 0 {
			return errors.New("counter delta must be positive")
		}
		MetricStorage.CounterInc(ctx, m.ID, *m.Delta)

	case "gauge":
		MetricStorage.GaugeSet(ctx, m.ID, *m.Value)

	default:
		return errors.New("invalid m type")
	}

	return nil
}
