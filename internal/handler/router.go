package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/HMasataka/gate/internal/middleware"
)

func NewRouter(
	healthHandler *HealthHandler,
	mw *middleware.Middleware,
) chi.Router {
	r := chi.NewRouter()

	// グローバルミドルウェア (順序重要)
	r.Use(mw.Recovery)
	r.Use(mw.RequestID)
	r.Use(mw.CORS)
	r.Use(mw.Logging)
	r.Use(mw.Metrics)

	// ヘルスチェック
	r.Get("/health", healthHandler.Liveness)
	r.Get("/ready", healthHandler.Readiness)

	// Prometheus メトリクス
	r.Handle("/metrics", promhttp.Handler())

	// 404 ハンドラ
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		Error(w, http.StatusNotFound, "not_found", "endpoint not found")
	})

	// 405 ハンドラ
	r.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		Error(w, http.StatusMethodNotAllowed, "method_not_allowed", "method not allowed")
	})

	return r
}
