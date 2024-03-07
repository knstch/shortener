package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	certconstructor "github.com/knstch/shortener/internal/app/certConstructor"
	"github.com/knstch/shortener/internal/app/handler"
	router "github.com/knstch/shortener/internal/app/router"
	dbconnect "github.com/knstch/shortener/internal/app/storage/DBConnect"
	memory "github.com/knstch/shortener/internal/app/storage/memory"
	"github.com/knstch/shortener/internal/app/storage/psql"

	"github.com/knstch/shortener/cmd/config"
	"github.com/knstch/shortener/internal/app/logger"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	config.ParseConfig()
	var storage handler.Storager
	var ping handler.PingChecker
	if config.ReadyConfig.DSN != "" {
		db, err := sql.Open("pgx", config.ReadyConfig.DSN)
		if err != nil {
			logger.ErrorLogger("Can't open connection: ", err)
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

	fmt.Printf("version=%s, time=%s, commit=%s\n", buildVersion, buildDate, buildCommit)

	srv := http.Server{
		Addr:    config.ReadyConfig.ServerAddr,
		Handler: router.RequestsRouter(h),
	}

	if config.ReadyConfig.EnableHTTPS {
		srv = http.Server{
			Addr:      config.ReadyConfig.ServerAddr,
			Handler:   router.RequestsRouter(h),
			TLSConfig: certconstructor.NewCert("shortener.ru").TLSConfig(),
		}
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.ErrorLogger("Shutdown error", err)
		}
		close(idleConnsClosed)
	}()

	switch {
	case config.ReadyConfig.EnableHTTPS:
		if err := srv.ListenAndServeTLS(config.ReadyConfig.CertFilePath, config.ReadyConfig.KeyFilePath); err != http.ErrServerClosed {
			logger.ErrorLogger("Run error", err)
		}
	case !config.ReadyConfig.EnableHTTPS:
		if err := srv.ListenAndServe(); err != http.ErrServerClosed {
			logger.ErrorLogger("Run error", err)
		}
	}
	<-idleConnsClosed
}
