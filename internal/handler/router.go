package handler

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/middleware"
)

func NewRouter(
	healthHandler *HealthHandler,
	authHandler *AuthHandler,
	oauthHandler *OAuthHandler,
	mfaHandler *MFAHandler,
	clientHandler *ClientHandler,
	adminClientHandler *AdminClientHandler,
	adminRoleHandler *AdminRoleHandler,
	adminUserHandler *AdminUserHandler,
	oidcHandler *OIDCHandler,
	socialHandler *SocialHandler,
	jwtManager domain.JWTManager,
	mw *middleware.Middleware,
	limiter domain.RateLimiter,
	httpsEnabled bool,
) chi.Router {
	r := chi.NewRouter()

	// グローバルミドルウェア (順序重要)
	r.Use(middleware.HTTPSRedirect(httpsEnabled))
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

	// JWKS エンドポイント
	r.Get("/.well-known/jwks.json", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, jwtManager.JWKS())
	})

	// OIDC Discovery エンドポイント
	r.Get("/.well-known/openid-configuration", oidcHandler.Discovery)

	// OIDC UserInfo エンドポイント (JWT 認証必須)
	r.Group(func(r chi.Router) {
		r.Use(middleware.JWTAuth(jwtManager))
		r.Get("/oauth/userinfo", oidcHandler.UserInfo)
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			// ログイン・登録: 10 req/min per IP
			r.With(middleware.RateLimit(limiter, 10, time.Minute)).Post("/register", authHandler.Register)
			r.With(middleware.RateLimit(limiter, 10, time.Minute)).Post("/login", authHandler.Login)
			r.Post("/logout", authHandler.Logout)
			r.Post("/verify-email", authHandler.VerifyEmail)
			r.Post("/resend-verification", authHandler.ResendVerification)
			// パスワードリセット: 3 req/min per IP
			r.With(middleware.RateLimit(limiter, 3, time.Minute)).Post("/forgot-password", authHandler.ForgotPassword)
			r.With(middleware.RateLimit(limiter, 3, time.Minute)).Post("/reset-password", authHandler.ResetPassword)

			// JWT 認証が必要なエンドポイント
			r.Group(func(r chi.Router) {
				r.Use(middleware.JWTAuth(jwtManager))
				r.Post("/change-password", authHandler.ChangePassword)
			})

			// ソーシャルログインエンドポイント
			r.Get("/social/{provider}/authorize", socialHandler.Authorize)
			r.Get("/social/{provider}/callback", socialHandler.Callback)
		})

		// OAuth 2.0 エンドポイント
		r.Route("/oauth", func(r chi.Router) {
			r.Get("/authorize", oauthHandler.Authorize)
			// OAuth トークン: 20 req/min per IP
			r.With(middleware.RateLimit(limiter, 20, time.Minute)).Post("/token", oauthHandler.Token)
			r.Post("/revoke", oauthHandler.Revoke)
			r.Post("/introspect", oauthHandler.Introspect)
		})

		// MFA エンドポイント (JWT 認証必須)
		r.Route("/mfa", func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtManager))
			r.Post("/totp/setup", mfaHandler.SetupTOTP)
			r.Post("/totp/confirm", mfaHandler.ConfirmTOTP)
			r.Delete("/totp", mfaHandler.DisableTOTP)
			r.Post("/recovery-codes/regenerate", mfaHandler.RegenerateRecoveryCodes)
		})

		// ユーザー向けクライアント管理 (JWT 認証必須)
		r.Route("/clients", func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtManager))
			r.Get("/", clientHandler.List)
			r.Post("/", clientHandler.Create)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", clientHandler.Get)
				r.Put("/", clientHandler.Update)
				r.Delete("/", clientHandler.Delete)
				r.Post("/rotate-secret", clientHandler.RotateSecret)
			})
		})

		// 管理者エンドポイント (JWT 認証必須)
		r.Route("/admin", func(r chi.Router) {
			r.Use(middleware.JWTAuth(jwtManager))

			r.Route("/clients", func(r chi.Router) {
				r.Get("/", adminClientHandler.List)
				r.Post("/", adminClientHandler.Create)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", adminClientHandler.Get)
					r.Put("/", adminClientHandler.Update)
					r.Delete("/", adminClientHandler.Delete)
					r.Post("/rotate-secret", adminClientHandler.RotateSecret)
				})
			})

			r.Route("/roles", func(r chi.Router) {
				r.Get("/", adminRoleHandler.ListRoles)
				r.Post("/", adminRoleHandler.CreateRole)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", adminRoleHandler.GetRole)
					r.Put("/", adminRoleHandler.UpdateRole)
					r.Delete("/", adminRoleHandler.DeleteRole)
					r.Post("/users/{userID}", adminRoleHandler.AssignRoleToUser)
					r.Delete("/users/{userID}", adminRoleHandler.RemoveRoleFromUser)
					r.Post("/permissions/{permID}", adminRoleHandler.AssignPermissionToRole)
					r.Delete("/permissions/{permID}", adminRoleHandler.RemovePermissionFromRole)
				})
			})

			r.Route("/permissions", func(r chi.Router) {
				r.Get("/", adminRoleHandler.ListPermissions)
				r.Post("/", adminRoleHandler.CreatePermission)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", adminRoleHandler.GetPermission)
					r.Put("/", adminRoleHandler.UpdatePermission)
					r.Delete("/", adminRoleHandler.DeletePermission)
				})
			})

			r.Route("/users", func(r chi.Router) {
				r.Get("/", adminUserHandler.List)
				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", adminUserHandler.Get)
					r.Put("/", adminUserHandler.Update)
					r.Delete("/", adminUserHandler.Delete)
					r.Post("/lock", adminUserHandler.Lock)
					r.Post("/unlock", adminUserHandler.Unlock)
					r.Post("/reset-mfa", adminUserHandler.ResetMFA)
					r.Post("/revoke-tokens", adminUserHandler.RevokeAllTokens)
				})
			})

			r.Get("/audit-logs", adminUserHandler.ListAuditLogs)
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
