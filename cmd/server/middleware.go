package main

import (
	"go.uber.org/zap"
	"net/http"
	"slices"
	"strings"
	"time"
)

type responseData struct {
	status int
	size   int
}

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
	compressEnabledForTypeList := []string{
		"application/json",
		"text/html",
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ow := w

		if slices.Contains(compressEnabledForTypeList, r.Header.Get("Accept")) {
			// проверяем, что клиент умеет получать от сервера сжатые данные в формате gzip
			acceptEncoding := r.Header.Get("Accept-Encoding")
			supportsGzip := strings.Contains(acceptEncoding, "gzip")
			if supportsGzip {
				r.Header.Set("Content-Encoding", "gzip")
				// оборачиваем оригинальный http.ResponseWriter новым с поддержкой сжатия
				cw := newCompressWriter(w)
				// меняем оригинальный http.ResponseWriter на новый
				ow = cw
				// не забываем отправить клиенту все сжатые данные после завершения middleware
				defer cw.Close()
			}
		}

		if slices.Contains(compressEnabledForTypeList, r.Header.Get("Content-Type")) {
			// проверяем, что клиент отправил серверу сжатые данные в формате gzip
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")
			if sendsGzip {
				// оборачиваем тело запроса в io.Reader с поддержкой декомпрессии
				cr, err := newCompressReader(r.Body)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				// меняем тело запроса на новое
				r.Body = cr
				defer cr.Close()
			}
		}

		// передаём управление хендлеру
		h.ServeHTTP(ow, r)
	})
}
