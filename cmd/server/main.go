// Сервер для сбора метрик. Хранит и отображает метрики.
// Хранит метрики тольтко для одного компьютера.
package main

import (
	"context"
	"crypto/rsa"
	"database/sql"
	"fmt"
	"github.com/gam6itko/go-musthave-metrics/internal/rsautils"
	"github.com/gam6itko/go-musthave-metrics/internal/server/config"
	"github.com/gam6itko/go-musthave-metrics/internal/server/controller"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/database"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/fallback"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/file"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/memory"
	"github.com/gam6itko/go-musthave-metrics/internal/server/storage/retrible"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.uber.org/zap"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// TODO: bad Паучье чутьё подсказывает, что так делать плохо. Но у меня пока что нет идей как сделать хорошо.
var MetricStorage storage.IStorage
var Database *sql.DB
var _key string
var _rsaPrivateKey *rsa.PrivateKey

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func init() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	cfg := initConfig()

	if cfg.RSAPrivateKey != "" {
		_rsaPrivateKey = loadPrivateKey(cfg.RSAPrivateKey)
	}
	if cfg.SignKey != "" {
		_key = cfg.SignKey
	}

	tmpDB, err := sql.Open("pgx", cfg.DatabaseDSN)
	if err != nil {
		log.Fatal(err)
	}
	Database = tmpDB
	if err2 := database.InitSchema(Database); err2 != nil {
		log.Fatalf("Failed to initialize database. %s", err2)
	}

	fileStorage := newFileStorage(cfg)
	MetricStorage = fallback.NewStorage(
		retrible.NewStorage(
			database.NewStorage(Database),
			[]time.Duration{
				time.Second,
				time.Second * 2,
				time.Second * 5,
			},
		),
		fileStorage,
	)

	server := &http.Server{
		Addr:    cfg.Address,
		Handler: newRouter(),
	}

	go catchSignal(server)

	Log.Info("Starting server", zap.String("addr", cfg.Address))
	if err := server.ListenAndServe(); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		Log.Info(err.Error(), zap.String("event", "start server"))
	}

	// on server.stop
	if err := fileStorage.Save(context.TODO()); err != nil {
		Log.Error(err.Error(), zap.String("event", "metrics save"))
	}
	if err2 := fileStorage.Close(); err2 != nil {
		log.Printf("ERROR. failed to close fileStorage: %v", err2)
	}
}

func loadPrivateKey(path string) *rsa.PrivateKey {
	b, err := os.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return rsautils.BytesToPrivateKey(b)
}

func newRouter() chi.Router {
	r := chi.NewRouter()

	r.Use(requestLoggingMiddleware)
	r.Use(hashCheckMiddleware)
	r.Use(rsaDecodeMiddleware)
	r.Use(compressMiddleware)

	ctrl := controller.NewMetricsController(MetricStorage, Log)
	r.Get("/", ctrl.GetAllMetricsHandler)
	r.Get("/value/{type}/{name}", ctrl.GetValue)
	r.Post("/update/{type}/{name}/{value}", ctrl.PostUpdate)
	// json
	r.Post("/value/", ctrl.PostValueJSONHandler)
	r.Post("/update/", ctrl.PostUpdateJSONHandler)
	r.Post("/updates/", ctrl.PostUpdateBatchJSONHandler)
	// database
	r.Get("/ping", getPingHandler)

	r.Mount("/debug", middleware.Profiler())

	return r
}

func newFileStorage(cfg config.Config) *file.Storage {
	sync := cfg.StoreInterval == 0
	fs, err := file.NewStorage(
		memory.NewStorage(),
		cfg.StoreFile,
		sync,
	)
	if err != nil {
		Log.Fatal(err.Error())
	}

	if cfg.Restore {
		if err := fs.Load(); err != nil {
			Log.Fatal(err.Error())
		}
	}

	if !sync {
		// Сохраняем каждые N секунд, если нет флага SYNC
		go func() {
			ticker := time.NewTicker(time.Duration(cfg.StoreInterval) * time.Second)
			for range ticker.C {
				if err2 := fs.Save(context.TODO()); err != nil { // грязновато, по идее нужно делать какой-то bridge-saver
					log.Fatal("Failed to save file", err2)
				}
			}
		}()
	}

	return fs
}

func catchSignal(server *http.Server) {
	terminateSignals := make(chan os.Signal, 1)

	signal.Notify(terminateSignals, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT) //NOTE:: syscall.SIGKILL we cannot catch kill -9 as its force kill signal.

	s := <-terminateSignals
	Log.Info("Got one of stop signals, shutting down server gracefully", zap.String("signal", s.String()))
	// metrics save

	if err := server.Shutdown(context.Background()); err != nil {
		Log.Info("Error from shutdown", zap.String("error", err.Error()))
	}
}
