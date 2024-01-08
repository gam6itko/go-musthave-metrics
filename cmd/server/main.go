package main

import (
	"context"
	"flag"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/file"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var MetricStorage *file.Storage

func main() {
	var fsConfig = &file.Config{} //create from flags
	var bindAddr string

	if envVal, exists := os.LookupEnv("ADDRESS"); exists {
		bindAddr = envVal
	}

	bindAddrTmp := flag.String("a", "", "Net address host:port")
	file.FromFlags(fsConfig, flag.CommandLine)
	flag.Parse()

	if err := file.FromEnv(fsConfig); err != nil {
		panic(err)
	}

	if bindAddr == "" {
		if *bindAddrTmp != "" {
			bindAddr = *bindAddrTmp
		} else {
			bindAddr = "localhost:8080"
		}
	}

	// Сохраняем метрики по интервалу
	MetricStorage = newFileStorage(fsConfig)

	server := &http.Server{
		Addr:    bindAddr,
		Handler: newRouter(),
	}

	go catchSignal(server)

	Log.Info("Starting server", zap.String("addr", bindAddr))
	if err := server.ListenAndServe(); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		Log.Info(err.Error(), zap.String("event", "start server"))
	}
}

func newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(requestLoggingMiddleware)
	r.Use(compressMiddleware)

	r.Get("/", getAllMetricsHandler)
	r.Get("/value/{type}/{name}", getValueHandler)
	r.Post("/update/{type}/{name}/{value}", postUpdateHandler)
	// json
	r.Post("/value/", postValueJSONHandler)
	r.Post("/update/", postUpdateJSONHandler)

	return r
}

func newFileStorage(fsConfig *file.Config) *file.Storage {
	sync := fsConfig.StoreInterval == 0
	fs := file.NewStorage(
		memory.NewStorage(),
		fsConfig.FileStoragePath,
		sync,
	)

	if fsConfig.Restore {
		if err := fs.Load(); err != nil {
			panic(err)
		}
	}

	if !sync {
		// Сохраняем каждые N секунд, если нет флага SYNC
		go func() {
			ticker := time.NewTicker(time.Duration(fsConfig.StoreInterval) * time.Second)
			for range ticker.C {
				fs.Save() // грязновато, по идее нужно делать какой-то bridge-saver
			}
		}()
	}

	return fs
}

func catchSignal(server *http.Server) {
	terminateSignals := make(chan os.Signal, 1)

	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM) //NOTE:: syscall.SIGKILL we cannot catch kill -9 as its force kill signal.

	s := <-terminateSignals
	Log.Info("Got one of stop signals, shutting down server gracefully", zap.String("signal", s.String()))
	if err := MetricStorage.Save(); err != nil {
		Log.Error(err.Error(), zap.String("event", "metrics save"))
	}
	err := server.Shutdown(context.Background())
	Log.Info("Error from shutdown", zap.String("error", err.Error()))
}
