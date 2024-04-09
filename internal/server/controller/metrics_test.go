package controller

import (
	"bytes"
	"context"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/go-chi/chi/v5"
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

	t.Run("PostUpdate counter 400", func(t *testing.T) {
		// for chi
		ctx := context.Background()
		ctx = context.WithValue(ctx,
			chi.RouteCtxKey,
			&chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"type", "name", "value"},
					Values: []string{"counter", "counter1", "bad-counter-value"},
				},
			},
		)

		w := httptest.NewRecorder()
		// url-path в данном тесте ни на что не влияет
		r := httptest.NewRequest(http.MethodPost, "/update/counter/counter1/bad-counter-value", nil)
		ctrl.PostUpdate(w, r.WithContext(ctx))
		require.Equal(t, 400, w.Code)

		_, err := storage.CounterGet(nil, "counter1")
		require.Error(t, err)
		require.EqualError(t, err, "not found")
	})

	t.Run("PostUpdate counter 200", func(t *testing.T) {
		// for chi
		ctx := context.Background()
		ctx = context.WithValue(ctx,
			chi.RouteCtxKey,
			&chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"type", "name", "value"},
					Values: []string{"counter", "counter1", "1"},
				},
			},
		)

		w := httptest.NewRecorder()
		// url-path в данном тесте ни на что не влияет
		r := httptest.NewRequest(http.MethodPost, "/update/counter/counter1/1", nil)
		ctrl.PostUpdate(w, r.WithContext(ctx))
		require.Equal(t, 200, w.Code)

		c, err := storage.CounterGet(nil, "counter1")
		require.NoError(t, err)
		require.Equal(t, int64(1), c)
		require.Equal(t, "OK", w.Body.String())
	})

	t.Run("PostUpdate gauge 200", func(t *testing.T) {
		// for chi
		ctx := context.Background()
		ctx = context.WithValue(ctx,
			chi.RouteCtxKey,
			&chi.Context{
				URLParams: chi.RouteParams{
					Keys:   []string{"type", "name", "value"},
					Values: []string{"gauge", "gauge1", "19.17"},
				},
			},
		)

		w := httptest.NewRecorder()
		r := httptest.NewRequest(http.MethodPost, "/update/gauge1/foo/19.17", nil)
		ctrl.PostUpdate(w, r.WithContext(ctx))
		require.Equal(t, 200, w.Code)

		g, err := storage.GaugeGet(nil, "gauge1")
		require.NoError(t, err)
		require.Equal(t, 19.17, g)
		require.Equal(t, "OK", w.Body.String())
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
