package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"

	"github.com/go-chi/chi/v5"

	config "github.com/knstch/shortener/cmd/config"
	URLstorage "github.com/knstch/shortener/internal/app/URLstorage"
	getShortenLink "github.com/knstch/shortener/internal/app/getShortenLink"
	logger "github.com/knstch/shortener/internal/app/logger"
	postLongLink "github.com/knstch/shortener/internal/app/postLongLink"
)

// Вызываем для передачи данных в функцию getURL
// и написания ответа в зависимости от ответа getURL
func getURL(res http.ResponseWriter, req *http.Request) {
	url := chi.URLParam(req, "url")
	if shortenURL := getShortenLink.GetShortenLink(url, URLstorage.StorageURLs); shortenURL != "" {
		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Location", shortenURL)
		res.WriteHeader(307)
		res.Write([]byte(shortenURL))
	} else {
		http.Error(res, "Bad Request", http.StatusBadRequest)
	}
}

// Вызывается при использовании метода POST, передает данные
// в функцию postURL для записи данных в хранилище и пишет
// ответ сервера, когда все записано
func postURL(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(201)
	res.Write([]byte(postLongLink.PostLongLink(string(body), &URLstorage.StorageURLs, config.ReadyConfig.BaseURL)))
}

// Роутер запросов
func RequestsRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/{url}", logger.RequestsLogger(getURL))
	r.Post("/", logger.RequestsLogger(postURL))
	return r
}

func main() {
	config.ParseConfig()
	srv := http.Server{
		Addr:    config.ReadyConfig.ServerAddr,
		Handler: RequestsRouter(),
	}

	idleConnsClosed := make(chan struct{})
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint

		if err := srv.Shutdown(context.Background()); err != nil {
			logger.ServerShutDownLog(err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.ServerRuns(err)
	}
	<-idleConnsClosed
}
