package controller

import (
	"bytes"
	"context"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMetricsController(t *testing.T) {
	logger := zaptest.NewLogger(t)
	storage := memory.NewStorage()
	ctrl := MetricsController{
		storage,
		logger,
	}

	t.Run("PostUpdate", func(t *testing.T) {
		ctx := context.WithValue(context.TODO(), "type", "gauge")
		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/update/gauge/foo/19.17", nil)
		r.WithContext(ctx)
		ctrl.PostUpdate(w, r)
		require.Equal(t, 200, w.Code)

		g, err := storage.GaugeGet(nil, "foo")
		require.NoError(t, err)
		require.Equal(t, 19.17, g)
	})
}

// Если несколько раз вызвать метод, то ссылки должны быть на разные области памяти.
func Test_decodeJsonRequest_metricNotSameRef(t *testing.T) {
	req, err := http.NewRequest(
		"POST",
		"/update",
		bytes.NewBufferString(`{"id": "foo", "type": "counter"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)
	m1, err := decodeMetricsRequest(req)
	require.NoError(t, err)
	req = nil

	req2, err := http.NewRequest(
		"POST",
		"/update",
		bytes.NewBufferString(`{"id": "foo", "type": "counter"}`),
	)
	require.NotNil(t, req2)

	req2.Header.Set("Content-Type", "application/json")
	require.NoError(t, err)
	m2, err := decodeMetricsRequest(req2)
	require.NoError(t, err)
	req2 = nil

	require.NotSame(t, *m1, *m2)
}

func Test_decodeMetricsBatchRequest(t *testing.T) {
	reqBody := `[
  {
    "id": "PollCount",
    "type": "counter",
    "delta": 1
  },
  {
    "id": "GaugeABC",
    "type": "gauge",
    "value": 19.17
  }
]`
	req := httptest.NewRequest("GET", "http://example.com/updates", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	list, err := decodeMetricsBatchRequest(req)
	require.NoError(t, err)
	require.Len(t, list, 2)
	// 0
	require.Equal(t, "PollCount", list[0].ID)
	require.Equal(t, "counter", list[0].MType)
	require.Equal(t, int64(1), *list[0].Delta)
	// 1
	require.Equal(t, "GaugeABC", list[1].ID)
	require.Equal(t, "gauge", list[1].MType)
	require.Equal(t, float64(19.17), *list[1].Value)
}
