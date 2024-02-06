package router

import (
	"github.com/go-chi/chi"
	"github.com/knstch/shortener/internal/app/handler"
	gzipCompressor "github.com/knstch/shortener/internal/app/middleware/gzipCompressor"
	loggerMiddleware "github.com/knstch/shortener/internal/app/middleware/loggerMiddleware"
)

func RequestsRouter(h *handler.Handler) chi.Router {
	router := chi.NewRouter()
	router.Use(gzipCompressor.GzipMiddleware)
	router.Use(loggerMiddleware.RequestsLogger)
	router.Get("/{url}", h.GetURL)
	router.Post("/", h.PostURL)
	router.Post("/api/shorten", h.PostLongLinkJSON)
	router.Get("/ping", h.PingDB)
	router.Post("/api/shorten/batch", h.PostBatch)
	router.Get("/api/user/urls", h.GetUserLinks)
	router.Delete("/api/user/urls", h.DeleteLinks)
	return router
}
