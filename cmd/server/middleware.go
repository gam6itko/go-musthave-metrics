package main

import (
	"bytes"
	"crypto/hmac"
	cryrand "crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	"go.uber.org/zap"
	"io"
	"log"
	"net"
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

func hashCheckMiddleware(handler http.Handler) http.Handler {
	if Cfg.SignKey == "" {
		return handler
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base64str := r.Header.Get("HashSHA256")
		if base64str == "" {
			handler.ServeHTTP(w, r)
			return
		}

		hash, err := base64.StdEncoding.DecodeString(base64str)
		if err != nil {
			Log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		bRequestBody, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err2 := r.Body.Close(); err2 != nil {
			Log.Error(err2.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		h := hmac.New(sha256.New, []byte(Cfg.SignKey))
		if _, err := h.Write(bRequestBody); err != nil {
			Log.Error(err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		dst := h.Sum(nil)
		if !bytes.Equal(hash, dst) {
			w.WriteHeader(http.StatusBadRequest)
			Log.Debug("Request sign mismatch")
			return
		}

		r.Body = io.NopCloser(bytes.NewBuffer(bRequestBody))
		handler.ServeHTTP(w, r)

		//todo w.Header().Set("HashSHA256", "response body hash")
	})
}

func rsaDecodeMiddleware(handler http.Handler) http.Handler {
	if Cfg.RSAPrivateKey == "" {
		return handler
	}

	privateKey := loadPrivateKey(Cfg.RSAPrivateKey)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		bRequestBody, err := io.ReadAll(r.Body)
		if err != nil {
			Log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		hash := sha256.New()
		b, err := rsautils.DecryptOAEP(hash, cryrand.Reader, privateKey, bRequestBody, nil)
		if err != nil {
			Log.Error(err.Error())
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err2 := r.Body.Close(); err2 != nil {
			Log.Error(err2.Error())
		}

		r.Body = io.NopCloser(bytes.NewBuffer(b))
		handler.ServeHTTP(w, r)

		if _, err2 := w.Write(b); err2 != nil {
			Log.Error(err2.Error())
		}
	})
}

func trustedSubnetMiddleware(handler http.Handler) http.Handler {
	if Cfg.TrustedSubnet == "" {
		return handler
	}

	_, subnetTrust, err := net.ParseCIDR(Cfg.TrustedSubnet)
	if err != nil {
		log.Fatal(err)
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ipStr := r.Header.Get("X-Real-IP")
		if ipStr == "" {
			w.WriteHeader(http.StatusForbidden)
			if _, err = fmt.Fprint(w, "X-Real-IP is required"); err != nil {
				log.Printf("ERROR. %s", err)
			}
			return
		}
		ip := net.ParseIP(ipStr)
		if ip == nil {
			w.WriteHeader(http.StatusForbidden)
			if _, err = fmt.Fprint(w, "X-Real-IP is incorrect"); err != nil {
				log.Printf("ERROR. %s", err)
			}
			return
		}

		if !subnetTrust.Contains(ip) {
			w.WriteHeader(http.StatusForbidden)
			if _, err = fmt.Fprint(w, "X-Real-IP not from trusted subnet"); err != nil {
				log.Printf("ERROR. %s", err)
			}
			return
		}

		handler.ServeHTTP(w, r)
	})
}
