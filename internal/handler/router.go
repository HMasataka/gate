package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/middleware"
)

func NewRouter(
	healthHandler *HealthHandler,
	authHandler *AuthHandler,
	oauthHandler *OAuthHandler,
	jwtManager domain.JWTManager,
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

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Post("/resend-verification", authHandler.ResendVerification)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)

			// JWT 認証が必要なエンドポイント
			r.Group(func(r chi.Router) {
				r.Use(middleware.JWTAuth(jwtManager))
				r.Post("/change-password", authHandler.ChangePassword)
			})
		})

		// OAuth 2.0 トークンエンドポイント
		r.Route("/oauth", func(r chi.Router) {
			r.Post("/token", oauthHandler.Token)
			r.Post("/revoke", oauthHandler.Revoke)
		})
	})

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
