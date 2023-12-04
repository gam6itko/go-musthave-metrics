package update

import (
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
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
			req := httptest.NewRequest(tt.method, tt.urlPath, nil)
			resp := httptest.NewRecorder()
			Handle(resp, req)

			res := resp.Result()
			assert.Equal(t, tt.want.code, res.StatusCode)
			defer res.Body.Close()
		})
	}
}
