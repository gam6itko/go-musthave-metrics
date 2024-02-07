package main

import (
	"context"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/file"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRequest(
	t *testing.T,
	ts *httptest.Server,
	method,
	path string,
) (*http.Response, string) {
	req, err := http.NewRequest(method, ts.URL+path, nil)
	require.NoError(t, err)

	resp, err := ts.Client().Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	require.NoError(t, err)

	return resp, string(respBody)
}

func TestPostUpdate(t *testing.T) {
	//todo-refactor Не уверен что хорошая практика инициализировать глобальную переменную в тестах
	var err error
	MetricStorage, err = file.NewStorage(memory.NewStorage(), "/tmp/tmp.json", false)
	require.NoError(t, err)

	ts := httptest.NewServer(newRouter())
	defer ts.Close()

	type want struct {
		code int
	}
	tests := []struct {
		name    string
		method  string
		urlPath string
		want    want
	}{
		{
			name:    "it work",
			method:  http.MethodPost,
			urlPath: "/update/gauge/Alloc/123",
			want: want{
				code: http.StatusOK,
			},
		},
		{
			name:    "incorrect type",
			method:  http.MethodPost,
			urlPath: "/update/wtf/wtf/123",
			want: want{
				code: http.StatusBadRequest,
			},
		},
		{
			name:    "not found",
			method:  http.MethodPost,
			urlPath: "/update/gauge/wtf",
			want: want{
				code: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := testRequest(t, ts, tt.method, tt.urlPath)
			defer result.Body.Close()
			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}

func TestGetValue(t *testing.T) {
	var err error
	MetricStorage = memory.NewStorage()
	require.NoError(t, err)

	ctx := context.Background()
	// preset
	MetricStorage.CounterInc(ctx, "fooCounter", 1)
	MetricStorage.CounterInc(ctx, "bar_c", 2)

	MetricStorage.GaugeSet(ctx, "foo_g", 1.1)
	MetricStorage.GaugeSet(ctx, "bar_g", 2.2)

	ts := httptest.NewServer(newRouter())
	defer ts.Close()

	type want struct {
		code  int
		value any
	}
	tests := []struct {
		name    string
		method  string
		urlPath string
		want    want
	}{
		// counter
		{
			name:    "counter foo",
			method:  http.MethodGet,
			urlPath: "/value/counter/fooCounter",
			want: want{
				code:  http.StatusOK,
				value: 1,
			},
		},
		{
			name:    "counter bar",
			method:  http.MethodGet,
			urlPath: "/value/counter/bar_c",
			want: want{
				code:  http.StatusOK,
				value: 2,
			},
		},
		// gauge
		{
			name:    "counter foo",
			method:  http.MethodGet,
			urlPath: "/value/gauge/foo_g",
			want: want{
				code:  http.StatusOK,
				value: 1.1,
			},
		},
		{
			name:    "counter bar",
			method:  http.MethodGet,
			urlPath: "/value/gauge/bar_g",
			want: want{
				code:  http.StatusOK,
				value: 2.2,
			},
		},

		// incorrect metric
		{
			name:    "incorrect type",
			method:  http.MethodGet,
			urlPath: "/value/wtf/wtf",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:    "gauge not exists",
			method:  http.MethodGet,
			urlPath: "/value/gauge/wtf",
			want: want{
				code: http.StatusNotFound,
			},
		},
		{
			name:    "counter not exists",
			method:  http.MethodGet,
			urlPath: "/value/counter/testSetGet123",
			want: want{
				code: http.StatusNotFound,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := testRequest(t, ts, tt.method, tt.urlPath)
			defer result.Body.Close()
			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}
