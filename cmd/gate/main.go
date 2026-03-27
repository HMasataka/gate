package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HMasataka/gate/internal/config"
	"github.com/HMasataka/gate/internal/domain"
	"github.com/HMasataka/gate/internal/handler"
	"github.com/HMasataka/gate/internal/infra/crypto"
	"github.com/HMasataka/gate/internal/infra/mailer"
	"github.com/HMasataka/gate/internal/infra/postgres"
	redisclient "github.com/HMasataka/gate/internal/infra/redis"
	"github.com/HMasataka/gate/internal/infra/social"
	"github.com/HMasataka/gate/internal/middleware"
	"github.com/HMasataka/gate/internal/usecase"
)


func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// 1. 設定読み込み
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// 2. 設定バリデーション
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	// 3. slog セットアップ
	setupLogger(cfg.Log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// 4. DB 接続
	db, err := postgres.NewDB(ctx, cfg.Database)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer db.Close()

	// 5. マイグレーション実行
	if err := postgres.RunMigrations(ctx, db, cfg.Database.MigrateTimeout); err != nil {
		return fmt.Errorf("run migrations: %w", err)
	}
	slog.Info("database migrations completed")

	// 6. Redis 接続
	rdb, err := redisclient.NewClient(ctx, cfg.Redis)
	if err != nil {
		return fmt.Errorf("connect redis: %w", err)
	}
	defer rdb.Close()

	// 7. インフラ層初期化
	hasher := crypto.NewArgon2Hasher(cfg.Argon2)
	random := &crypto.SecureRandom{}
	sessionStore := redisclient.NewSessionStore(rdb, cfg.Session)
	mail := mailer.NewStdoutMailer()
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewRefreshTokenRepo(db)
	clientRepo := postgres.NewClientRepo(db)
	roleRepo := postgres.NewRoleRepo(db)
	permRepo := postgres.NewPermissionRepo(db)
	socialRepo := postgres.NewSocialConnectionRepo(db)

	jwtManager, err := crypto.NewJWTManager(cfg.JWT, cfg.Token)
	if err != nil {
		return fmt.Errorf("create jwt manager: %w", err)
	}

	// 8. ユースケース初期化
	tokenUsecase := usecase.NewTokenUsecase(tokenRepo, userRepo, jwtManager, random, cfg.Token)
	authUsecase := usecase.NewAuthUsecase(userRepo, hasher, mail, sessionStore, random, tokenUsecase, cfg.Auth, cfg.Session)
	mfaUsecase := usecase.NewMFAUsecase(userRepo, random, cfg.MFA)
	// AuthorizationCodeRepository は未実装のため nil を渡す
	var codeRepo domain.AuthorizationCodeRepository
	oauthUsecase := usecase.NewOAuthUsecase(clientRepo, codeRepo, tokenUsecase, random, cfg.OAuth, cfg.Token)
	clientUsecase := usecase.NewClientUsecase(clientRepo, random, cfg.OAuth)
	permCache := redisclient.NewPermissionCache(rdb, permRepo, 5*time.Minute)
	roleUsecase := usecase.NewRoleUsecase(roleRepo, permRepo, permCache, random)
	permUsecase := usecase.NewPermissionUsecase(permRepo, permCache, random)
	serverURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	oidcUsecase := usecase.NewOIDCUsecase(userRepo, jwtManager, serverURL)

	// ソーシャルプロバイダ設定
	socialProviders := map[string]usecase.SocialProvider{}
	if cfg.Social.GoogleClientID != "" {
		googleProvider := social.NewGoogleProvider(cfg.Social.GoogleClientID, cfg.Social.GoogleClientSecret, cfg.Social.GoogleRedirectURI)
		socialProviders["google"] = social.NewProviderAdapter(googleProvider)
	}
	if cfg.Social.GitHubClientID != "" {
		githubProvider := social.NewGitHubProvider(cfg.Social.GitHubClientID, cfg.Social.GitHubClientSecret, cfg.Social.GitHubRedirectURI)
		socialProviders["github"] = social.NewProviderAdapter(githubProvider)
	}
	socialUsecase := usecase.NewSocialUsecase(socialRepo, userRepo, sessionStore, tokenUsecase, random, socialProviders, cfg.Session)

	// 9. ミドルウェア初期化
	mw := middleware.New(cfg)

	// 10. ハンドラ初期化
	healthHandler := handler.NewHealthHandler(db, rdb)
	authHandler := handler.NewAuthHandler(authUsecase)
	oauthHandler := handler.NewOAuthHandler(oauthUsecase, tokenUsecase)
	mfaHandler := handler.NewMFAHandler(mfaUsecase, hasher)
	adminClientHandler := handler.NewAdminClientHandler(clientUsecase)
	adminRoleHandler := handler.NewAdminRoleHandler(roleUsecase, permUsecase)
	oidcHandler := handler.NewOIDCHandler(oidcUsecase)
	socialHandler := handler.NewSocialHandler(socialUsecase)

	// 11. ルーター構築
	router := handler.NewRouter(healthHandler, authHandler, oauthHandler, mfaHandler, adminClientHandler, adminRoleHandler, oidcHandler, socialHandler, jwtManager, mw)

	// 12. HTTP サーバー起動
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           router,
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReadHeaderTimeout: cfg.Server.ReadHeaderTimeout,
	}

	errCh := make(chan error, 1)
	go func() {
		slog.Info("starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
		close(errCh)
	}()

	// 13. グレースフルシャットダウン
	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
		slog.Info("shutting down server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	slog.Info("server stopped gracefully")
	return nil
}

func setupLogger(cfg config.LogConfig) {
	var level slog.Level
	switch cfg.Level {
	case "DEBUG":
		level = slog.LevelDebug
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{Level: level}

	var h slog.Handler
	if cfg.Format == "text" {
		h = slog.NewTextHandler(os.Stdout, opts)
	} else {
		h = slog.NewJSONHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(h))
}
