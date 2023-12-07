package main

import (
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
			assert.Equal(t, tt.want.code, result.StatusCode)
		})
	}
}
