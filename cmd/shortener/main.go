package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"
	"github.com/knstch/shortener/internal/app/handler"
	dbconnect "github.com/knstch/shortener/internal/app/storage/DBConnect"
	memory "github.com/knstch/shortener/internal/app/storage/memory"
	"github.com/knstch/shortener/internal/app/storage/psql"

	"github.com/knstch/shortener/cmd/config"
	"github.com/knstch/shortener/internal/app/logger"
	gzipCompressor "github.com/knstch/shortener/internal/app/middleware/gzipCompressor"
	loggerMiddleware "github.com/knstch/shortener/internal/app/middleware/loggerMiddleware"
)

// Роутер запросов
func RequestsRouter(h *handler.Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(gzipCompressor.GzipMiddleware)
	router.Use(loggerMiddleware.RequestsLogger)
	router.Get("/{url}", h.GetURL)
	router.Post("/", h.PostURL)
	router.Post("/api/shorten", h.PostLongLinkJSON)
	router.Get("/ping", h.PingDB)
	router.Post("/api/shorten/batch", h.PostBatch)
	return router
}

func main() {
	config.ParseConfig()
	// storage.StorageURLs.Load(config.ReadyConfig.FileStorage)
	var storage handler.Storage
	var ping handler.PingChecker
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {

		}
		err = psql.InitDB(db)
		if err != nil {
			logger.ErrorLogger("Can't init DB: ", err)
		}
		storage = psql.NewPsqlStorage(db)
		ping = dbconnect.NewDBConnection(db)
	} else {
		storage = memory.NewMemStorage()
	}
	h := handler.NewHandler(storage, ping)

	srv := http.Server{
		Addr:    config.ReadyConfig.ServerAddr,
		Handler: RequestsRouter(h),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.ErrorLogger("Shutdown error", err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.ErrorLogger("Run error", err)
	}
	<-idleConnsClosed
}
