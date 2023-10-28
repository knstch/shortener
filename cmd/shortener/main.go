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

	gzipCompressor "github.com/knstch/shortener/internal/app/gzipCompressor"
	logger "github.com/knstch/shortener/internal/app/logger"
	postLongLink "github.com/knstch/shortener/internal/app/postLongLink"
	postLongLinkJSON "github.com/knstch/shortener/internal/app/postLongLinkJSON"
)

// Вызываем для передачи данных в функцию getURL
// и написания ответа в зависимости от ответа getURL
func getURL(res http.ResponseWriter, req *http.Request) {
	url := chi.URLParam(req, "url")
	if shortenURL := getShortenLink.GetShortenLink(url, URLstorage.StorageURLs); shortenURL != "" {
		res.Header().Set("Content-Type", "text/plain")
		res.Header().Set("Content-Encoding", "gzip")
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
		logger.ErrorLogger("Error during reading body: ", err)
	}
	res.Header().Set("Content-Type", "text/plain")
	res.WriteHeader(201)
	res.Write([]byte(postLongLink.PostLongLink(string(body), &URLstorage.StorageURLs, config.ReadyConfig.BaseURL)))
}

// Передаем json-объект и получаем в ответе короткий URL в виде json-объекта
func postURLJSON(res http.ResponseWriter, req *http.Request) {
	body, err := io.ReadAll(req.Body)
	if err != nil {
		logger.ErrorLogger("Error during opening body: ", err)
	}
	res.Header().Set("Content-Type", "application/json")
	res.WriteHeader(201)
	res.Write([]byte(postLongLinkJSON.PostLongLinkJSON(body)))
}

// Роутер запросов
func RequestsRouter() chi.Router {
	router := chi.NewRouter()
	router.Use(logger.RequestsLogger)
	router.Use(gzipCompressor.GzipMiddleware)
	router.Get("/{url}", getURL)
	router.Post("/", postURL)
	router.Post("/api/shorten", postURLJSON)
	return router
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
			logger.ErrorLogger("Shutdown error", err)
		}
		close(idleConnsClosed)
	}()
	if err := srv.ListenAndServe(); err != http.ErrServerClosed {
		logger.ErrorLogger("Run error", err)
	}
	<-idleConnsClosed
}
