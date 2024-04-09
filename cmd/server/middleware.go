package main

import (
	"go.uber.org/zap"
	"net/http"
	"strings"
	"time"
)

var (
	_ http.ResponseWriter = (*loggingResponseWriter)(nil)
)

type responseData struct {
	status int
	size   int
}

// loggingResponseWriter логирует входящие запросы.
type loggingResponseWriter struct {
	http.ResponseWriter
	responseData *responseData
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b)
	r.responseData.size += size // захватываем размер
	return size, err
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode)
	r.responseData.status = statusCode // захватываем код статуса
}

func requestLoggingMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		uri := r.RequestURI
		method := r.Method

		rd := &responseData{
			status: 0,
			size:   0,
		}
		lw := loggingResponseWriter{
			ResponseWriter: w,
			responseData:   rd,
		}

		h.ServeHTTP(&lw, r)

		duration := time.Since(start)

		Log.Info(
			"Request",
			zap.String("uri", uri),
			zap.String("method", method),
			zap.Duration("duration", duration),
			// response
			zap.Int("status", rd.status),
			zap.Int("size", rd.size),
		)
	})
}

func compressMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if contentEncoding := r.Header.Get("Content-Encoding"); strings.Contains(contentEncoding, "gzip") {
			cr, err := newCompressReader(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			// меняем тело запроса на новое
			r.Body = cr
			defer cr.Close()
		}

		if acceptEncoding := r.Header.Get("Accept-Encoding"); strings.Contains(acceptEncoding, "gzip") {
			cw := newCompressWriter(w)
			ow = cw
			defer cw.Close()
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}
