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

	t.Run("PostUpdate", func(t *testing.T) {
		t.Run("counter 400", func(t *testing.T) {
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

			_, err := storage.CounterGet(context.TODO(), "counter1")
			require.Error(t, err)
			require.EqualError(t, err, "not found")
		})

		t.Run("counter 200", func(t *testing.T) {
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

			c, err := storage.CounterGet(context.TODO(), "counter1")
			require.NoError(t, err)
			require.Equal(t, int64(1), c)
			require.Equal(t, "OK", w.Body.String())
		})

		t.Run("counter 400", func(t *testing.T) {
			// for chi
			ctx := context.Background()
			ctx = context.WithValue(ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"type", "name", "value"},
						Values: []string{"gauge", "gauge1", "bad-gauge-value"},
					},
				},
			)

			w := httptest.NewRecorder()
			// url-path в данном тесте ни на что не влияет
			r := httptest.NewRequest(http.MethodPost, "/update/gauge/gauge1/bad-gauge-value", nil)
			ctrl.PostUpdate(w, r.WithContext(ctx))
			require.Equal(t, 400, w.Code)

			_, err := storage.GaugeGet(context.TODO(), "gauge1")
			require.Error(t, err)
			require.EqualError(t, err, "not found")
		})

		t.Run("gauge 200", func(t *testing.T) {
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

			g, err := storage.GaugeGet(context.TODO(), "gauge1")
			require.NoError(t, err)
			require.InEpsilon(t, 19.17, g, 1)
			require.Equal(t, "OK", w.Body.String())
		})
	})

	t.Run("PostUpdateJSONHandler", func(t *testing.T) {
		t.Run("counter 200", func(t *testing.T) {
			json := `{
	"id": "counter2",
	"type": "counter",
	"delta": 2
}`
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer([]byte(json)))
			r.Header.Set("Content-Type", "application/json")
			ctrl.PostUpdateJSONHandler(w, r)
			require.Equal(t, 200, w.Code)
			require.Equal(t, "OK", w.Body.String())

			c, err := storage.CounterGet(context.TODO(), "counter2")
			require.NoError(t, err)
			require.Equal(t, int64(2), c)
		})

		t.Run("gauge 200", func(t *testing.T) {
			json := `{
	"id": "gauge2",
	"type": "gauge",
	"value": 19.22
}`
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/update/", bytes.NewBuffer([]byte(json)))
			r.Header.Set("Content-Type", "application/json")
			ctrl.PostUpdateJSONHandler(w, r)
			require.Equal(t, 200, w.Code)
			require.Equal(t, "OK", w.Body.String())

			g, err := storage.GaugeGet(context.TODO(), "gauge2")
			require.NoError(t, err)
			require.InEpsilon(t, 19.22, g, 1)
		})
	})

	t.Run("GetValue", func(t *testing.T) {
		t.Run("wrong type - 404", func(t *testing.T) {
			// for chi
			ctx := context.Background()
			ctx = context.WithValue(ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"type", "name"},
						Values: []string{"foo", "foo777"},
					},
				},
			)

			w := httptest.NewRecorder()
			// url-path в данном тесте ни на что не влияет
			r := httptest.NewRequest(http.MethodPost, "/value/foo/foo777", nil)
			ctrl.GetValue(w, r.WithContext(ctx))
			require.Equal(t, 404, w.Code)
		})

		t.Run("counter2 - 200", func(t *testing.T) {
			// for chi
			ctx := context.Background()
			ctx = context.WithValue(ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"type", "name"},
						Values: []string{"counter", "counter2"},
					},
				},
			)

			w := httptest.NewRecorder()
			// url-path в данном тесте ни на что не влияет
			r := httptest.NewRequest(http.MethodPost, "/value/counter/counter2", nil)
			ctrl.GetValue(w, r.WithContext(ctx))
			require.Equal(t, 200, w.Code)
			require.Equal(t, "2", w.Body.String())
		})

		t.Run("counter not found", func(t *testing.T) {
			// for chi
			ctx := context.Background()
			ctx = context.WithValue(ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"type", "name"},
						Values: []string{"counter", "counter666"},
					},
				},
			)

			w := httptest.NewRecorder()
			// url-path в данном тесте ни на что не влияет
			r := httptest.NewRequest(http.MethodPost, "/value/counter/counter666", nil)
			ctrl.GetValue(w, r.WithContext(ctx))
			require.Equal(t, 404, w.Code)
		})

		t.Run("gauge2 - 200", func(t *testing.T) {
			// for chi
			ctx := context.Background()
			ctx = context.WithValue(ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"type", "name"},
						Values: []string{"gauge", "gauge2"},
					},
				},
			)

			w := httptest.NewRecorder()
			// url-path в данном тесте ни на что не влияет
			r := httptest.NewRequest(http.MethodPost, "/value/gauge/gauge2", nil)
			ctrl.GetValue(w, r.WithContext(ctx))
			require.Equal(t, 200, w.Code)
			require.Equal(t, "19.22", w.Body.String())
		})

		t.Run("gauge not found", func(t *testing.T) {
			// for chi
			ctx := context.Background()
			ctx = context.WithValue(ctx,
				chi.RouteCtxKey,
				&chi.Context{
					URLParams: chi.RouteParams{
						Keys:   []string{"type", "name"},
						Values: []string{"gauge", "gauge666"},
					},
				},
			)

			w := httptest.NewRecorder()
			// url-path в данном тесте ни на что не влияет
			r := httptest.NewRequest(http.MethodPost, "/value/gauge/gauge666", nil)
			ctrl.GetValue(w, r.WithContext(ctx))
			require.Equal(t, 404, w.Code)
		})
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
	require.InDelta(t, float64(19.17), *list[1].Value, .1)
}
