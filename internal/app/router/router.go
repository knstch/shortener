// Модуль router отвечает за роутинг запросов и соединения с нужным хендлером.
package router

import (
	// pprof "net/http/pprof"

	"github.com/go-chi/chi"
	"github.com/knstch/shortener/internal/app/handler"

	gzipCompressor "github.com/knstch/shortener/internal/app/middleware/gzipCompressor"
	loggerMiddleware "github.com/knstch/shortener/internal/app/middleware/loggerMiddleware"
)

// RequestsRouter - это роутер.
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
	router.Get("/api/internal/stats", h.GetStatsHandler)
	// router.Get("/debug/pprof/", pprof.Index)
	// router.Get("/debug/pprof/heap", pprof.Index)
	// router.Get("/debug/pprof/cmdline", pprof.Cmdline)
	// router.Get("/debug/pprof/profile", pprof.Profile)
	// router.Get("/debug/pprof/symbol", pprof.Symbol)
	// router.Get("/debug/pprof/trace", pprof.Trace)
	return router
}
