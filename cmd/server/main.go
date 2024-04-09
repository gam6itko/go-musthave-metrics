package main

import (
	"context"
	"database/sql"
	"flag"
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
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// todo-bad Паучье чутьё подсказывает, что так делать плохо. Но у меня пока что нет идей как сделать хорошо.
var MetricStorage storage.Storage
var Database *sql.DB

// @Title Get All Metrics
// @Description Накопление и отображение метрик.
// @Version 1.0
// @Contact.email gam6itko@yandex.ru
// @BasePath /
// @Host localhost:8080

func main() {
	var fsConfig = &file.Config{} //create from flags
	var bindAddr string
	var dbDsn string

	if envVal, exists := os.LookupEnv("ADDRESS"); exists {
		bindAddr = envVal
	}

	//init db
	if envVal, exists := os.LookupEnv("DATABASE_DSN"); exists {
		dbDsn = envVal
	}

	bindAddrTmp := flag.String("a", "", "Net address host:port")
	dbDsnTmp := flag.String("d", "", "Database DSN")
	file.FromFlags(fsConfig, flag.CommandLine)
	flag.Parse()

	if err := file.FromEnv(fsConfig); err != nil {
		Log.Fatal(err.Error())
	}

	if bindAddr == "" {
		if *bindAddrTmp != "" {
			bindAddr = *bindAddrTmp
		} else {
			bindAddr = "localhost:8080"
		}
	}

	// database open
	if *dbDsnTmp != "" {
		dbDsn = *dbDsnTmp
	}

	tmpDB, err := sql.Open("pgx", dbDsn)
	if err != nil {
		panic(err)
	}
	Database = tmpDB
	database.InitSchema(Database)

	fileStorage := newFileStorage(fsConfig)
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
		Addr:    bindAddr,
		Handler: newRouter(),
	}

	go catchSignal(server)

	Log.Info("Starting server", zap.String("addr", bindAddr))
	if err := server.ListenAndServe(); err != nil {
		// записываем в лог ошибку, если сервер не запустился
		Log.Info(err.Error(), zap.String("event", "start server"))
	}

	// on server.stop
	if err := fileStorage.Save(); err != nil {
		Log.Error(err.Error(), zap.String("event", "metrics save"))
	}
	fileStorage.Close()
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
	r.Post("/updates/", postUpdateBatchJSONHandler)
	// database
	r.Get("/ping", getPingHandler)

	r.Mount("/debug", middleware.Profiler())

	return r
}

func newFileStorage(fsConfig *file.Config) *file.Storage {
	sync := fsConfig.StoreInterval == 0
	fs, err := file.NewStorage(
		memory.NewStorage(),
		fsConfig.FileStoragePath,
		sync,
	)
	if err != nil {
		Log.Fatal(err.Error())
	}

	if fsConfig.Restore {
		if err := fs.Load(); err != nil {
			Log.Fatal(err.Error())
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
	// metrics save

	err := server.Shutdown(context.Background())
	Log.Info("Error from shutdown", zap.String("error", err.Error()))
}
