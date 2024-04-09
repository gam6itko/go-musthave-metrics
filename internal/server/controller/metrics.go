package controller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/common"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"strings"
)

// MetricsController обрабатывает запросы связанные с метриками.
type MetricsController struct {
	storage storage.IStorage
	logger  *zap.Logger
}

func NewMetricsController(storage storage.IStorage, logger *zap.Logger) *MetricsController {
	return &MetricsController{storage, logger}
}

// GetAllMetricsHandler возвращает все накопленные метрики в формате HTML.
//
// Health godoc
// @Tags Info
// @Summary Получить все накопленные метрики в формате HTML.
// @ID GetAllMetrics
// @Produce text/html
// @Success 200 {string} string "Метрики"
// @Failure 500 {string} string "Внутренняя ошибка"
// @Router / [get]
func (ths MetricsController) GetAllMetricsHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "text/html") //iter8 fix

	io.WriteString(resp, "<h2>All metrics</h2>")

	io.WriteString(resp, "<h2>Counter</h2>")
	counterAll, _ := ths.storage.CounterAll(req.Context())
	for name, val := range counterAll {
		io.WriteString(resp, fmt.Sprintf("<div>%s: %d</div>", name, val))
	}

	io.WriteString(resp, "<h2>Gauge</h2>")
	gaugeAll, _ := ths.storage.GaugeAll(req.Context())
	for name, val := range gaugeAll {
		io.WriteString(resp, fmt.Sprintf("<div>%s: %f</div>", name, val))
	}
}

// GetValue возвращает одно значение метрики.
func (ths MetricsController) GetValue(resp http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	if name == "" {
		http.Error(resp, "Bad name", http.StatusNotFound)
		return
	}

	switch chi.URLParam(req, "type") {
	case "counter":
		val, err := ths.storage.CounterGet(req.Context(), name)
		if err != nil {
			http.Error(resp, "Not found", http.StatusNotFound)
			return
		}
		io.WriteString(resp, fmt.Sprintf("%d", val))

	case "gauge":
		val, err := ths.storage.GaugeGet(req.Context(), name)
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

// PostUpdate сохраняет одну метрику c помощью передачи параметров в url-path.
//
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
func (ths MetricsController) PostUpdate(resp http.ResponseWriter, req *http.Request) {
	name := chi.URLParam(req, "name")
	value := chi.URLParam(req, "value")

	switch strings.ToLower(chi.URLParam(req, "type")) {
	case "counter":
		v, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			http.Error(resp, "invalid counter value", http.StatusBadRequest)
			return
		}
		if err := ths.storage.CounterInc(req.Context(), name, v); err != nil {
			http.Error(resp, "fail to counter inc", http.StatusInternalServerError)
			return
		}

	case "gauge":
		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			http.Error(resp, "invalid gauge value", http.StatusBadRequest)
			return
		}
		if err := ths.storage.GaugeSet(req.Context(), name, v); err != nil {
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

// PostValueJSONHandler запрос на получение одной метрики.
func (ths MetricsController) PostValueJSONHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metric, err := decodeMetricsRequest(req)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		ths.logger.Warn(err.Error())
		return
	}

	switch metric.MType {
	case "counter":
		val, _ := ths.storage.CounterGet(req.Context(), metric.ID)
		metric.Delta = &val

	case "gauge":
		val, _ := ths.storage.GaugeGet(req.Context(), metric.ID)
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
		ths.logger.Error(err.Error())
	}
}

// PostUpdateJSONHandler обновляет одну метрику из запроса в формате JSON.
func (ths MetricsController) PostUpdateJSONHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metric, err := decodeMetricsRequest(req)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		ths.logger.Warn(err.Error())
		return
	}

	if err := ths.persistMetric(req.Context(), metric); err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		return
	}

	//b, err := json.Marshal(resp)
	//if err != nil {
	//	httpErrorJSON(resp, err.Error(), http.StatusInternalServerError)
	//	return
	//}

	resp.WriteHeader(http.StatusOK)
	_, err = resp.Write([]byte("OK"))
	if err != nil {
		ths.logger.Error(err.Error())
	}
}

// PostUpdateBatchJSONHandler обновляет несколько метрик за один запроса.
func (ths MetricsController) PostUpdateBatchJSONHandler(resp http.ResponseWriter, req *http.Request) {
	resp.Header().Set("Content-Type", "application/json")

	metricList, err := decodeMetricsBatchRequest(req)
	if err != nil {
		httpErrorJSON(resp, err.Error(), http.StatusBadRequest)
		ths.logger.Warn(err.Error())
		return
	}

	for _, m := range metricList {
		if err := ths.persistMetric(req.Context(), &m); err != nil {
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
		ths.logger.Error(err.Error())
	}
}

// decodeMetricsRequest превращает request-body в объект метрики.
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

// httpErrorJSON отправляет ошибку в формате JSON.
func httpErrorJSON(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	fmt.Fprintf(w, `{"error":"%s"}`, message)
}

// persistMetric Сохраняем метрику в хранилище.
func (ths MetricsController) persistMetric(ctx context.Context, m *common.Metrics) error {
	switch strings.ToLower(m.MType) {
	case "counter":
		if *m.Delta < 0 {
			return errors.New("counter delta must be positive")
		}
		ths.storage.CounterInc(ctx, m.ID, *m.Delta)

	case "gauge":
		ths.storage.GaugeSet(ctx, m.ID, *m.Value)

	default:
		return errors.New("invalid m type")
	}

	return nil
}
