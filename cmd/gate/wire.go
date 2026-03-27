package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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

type app struct {
	router       http.Handler
	auditUsecase *usecase.AuditUsecase
	userRepo     domain.UserRepository
	cleanup      func()
}

func initApp(ctx context.Context, cfg *config.Config) (*app, error) {
	db, err := postgres.NewDB(ctx, cfg.Database)
	if err != nil {
		return nil, fmt.Errorf("connect database: %w", err)
	}

	if err := postgres.RunMigrations(ctx, db, cfg.Database.MigrateTimeout); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}
	slog.Info("database migrations completed")

	rdb, err := redisclient.NewClient(ctx, cfg.Redis)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("connect redis: %w", err)
	}

	hasher := crypto.NewArgon2Hasher(cfg.Argon2)
	random := &crypto.SecureRandom{}
	sessionStore := redisclient.NewSessionStore(rdb, cfg.Session)
	rateLimiter := redisclient.NewRateLimiterStore(rdb)
	_ = crypto.NewJTIStore(rdb) // JTI replay prevention store (integration point for future use)
	mail := mailer.NewStdoutMailer()
	userRepo := postgres.NewUserRepo(db)
	tokenRepo := postgres.NewRefreshTokenRepo(db)
	clientRepo := postgres.NewClientRepo(db)
	roleRepo := postgres.NewRoleRepo(db)
	permRepo := postgres.NewPermissionRepo(db)
	socialRepo := postgres.NewSocialConnectionRepo(db)

	jwtManager, err := crypto.NewJWTManager(cfg.JWT, cfg.Token)
	if err != nil {
		rdb.Close()
		db.Close()
		return nil, fmt.Errorf("create jwt manager: %w", err)
	}

	txManager := postgres.NewTxManager(db)
	tokenUsecase := usecase.NewTokenUsecase(tokenRepo, userRepo, jwtManager, random, cfg.Token, txManager)
	authUsecase := usecase.NewAuthUsecase(userRepo, hasher, mail, sessionStore, random, tokenUsecase, cfg.Auth, cfg.Session)
	mfaUsecase := usecase.NewMFAUsecase(userRepo, random, cfg.MFA)
	// AuthorizationCodeRepository は未実装のため nil を渡す
	var codeRepo domain.AuthorizationCodeRepository
	oauthUsecase := usecase.NewOAuthUsecase(clientRepo, codeRepo, tokenUsecase, random, cfg.OAuth, cfg.Token, txManager)
	clientUsecase := usecase.NewClientUsecase(clientRepo, random, cfg.OAuth)
	permCache := redisclient.NewPermissionCache(rdb, permRepo, 5*time.Minute)
	roleUsecase := usecase.NewRoleUsecase(roleRepo, permRepo, permCache, random)
	permUsecase := usecase.NewPermissionUsecase(permRepo, permCache, random)
	serverURL := fmt.Sprintf("http://localhost:%d", cfg.Server.Port)
	oidcUsecase := usecase.NewOIDCUsecase(userRepo, jwtManager, serverURL)

	socialProviders := map[string]usecase.SocialProvider{}
	if cfg.Social.GoogleClientID != "" {
		googleProvider := social.NewGoogleProvider(cfg.Social.GoogleClientID, cfg.Social.GoogleClientSecret, cfg.Social.GoogleRedirectURI)
		socialProviders["google"] = social.NewProviderAdapter(googleProvider)
	}
	if cfg.Social.GitHubClientID != "" {
		githubProvider := social.NewGitHubProvider(cfg.Social.GitHubClientID, cfg.Social.GitHubClientSecret, cfg.Social.GitHubRedirectURI)
		socialProviders["github"] = social.NewProviderAdapter(githubProvider)
	}
	socialUsecase := usecase.NewSocialUsecase(socialRepo, userRepo, sessionStore, tokenUsecase, random, socialProviders, cfg.Session, txManager)
	auditRepo := postgres.NewAuditLogRepo(db)
	auditUsecase := usecase.NewAuditUsecase(auditRepo, random, cfg.Auth.AuditRetentionDays)
	userUsecase := usecase.NewUserUsecase(userRepo, sessionStore, tokenUsecase, random)

	mw := middleware.New(cfg)

	healthHandler := handler.NewHealthHandler(db, rdb)
	authHandler := handler.NewAuthHandler(authUsecase)
	oauthHandler := handler.NewOAuthHandler(oauthUsecase, tokenUsecase)
	mfaHandler := handler.NewMFAHandler(mfaUsecase, hasher)
	adminClientHandler := handler.NewAdminClientHandler(clientUsecase)
	adminRoleHandler := handler.NewAdminRoleHandler(roleUsecase, permUsecase)
	adminUserHandler := handler.NewAdminUserHandler(userUsecase, auditUsecase)
	oidcHandler := handler.NewOIDCHandler(oidcUsecase)
	socialHandler := handler.NewSocialHandler(socialUsecase)

	router := handler.NewRouter(healthHandler, authHandler, oauthHandler, mfaHandler, adminClientHandler, adminRoleHandler, adminUserHandler, oidcHandler, socialHandler, jwtManager, mw, rateLimiter, cfg.RateLimit.HTTPSRedirect)

	cleanup := func() {
		rdb.Close()
		db.Close()
	}

	return &app{
		router:       router,
		auditUsecase: auditUsecase,
		userRepo:     userRepo,
		cleanup:      cleanup,
	}, nil
}
