package main

import (
	"flag"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
)

func main() {
	bindAddr := "localhost:8080"

	if envVal := os.Getenv("ADDRESS"); envVal != "" {
		bindAddr = envVal
	}

	fBindAddrRef := flag.String("a", "", "Net address host:port")
	flag.Parse()

	if *fBindAddrRef != "" {
		bindAddr = *fBindAddrRef
	}

	Log.Info("Starting server", zap.String("addr", bindAddr))
	if err := http.ListenAndServe(bindAddr, newRouter()); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		Log.Fatal(err.Error(), zap.String("event", "start server"))
	}
}

func newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(requestLoggingMiddleware)

	r.Get("/", getAllMetricsHandler)
	r.Get("/value/{type}/{name}", getValueHandler)
	r.Post("/update/{type}/{name}/{value}", postUpdateHandler)
	// json
	r.Post("/value/", postValueJsonHandler)
	r.Post("/update/", postUpdateJsonHandler)

	return r
}
