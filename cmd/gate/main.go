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
	"github.com/HMasataka/gate/internal/usecase"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application error", "error", err)
		os.Exit(1)
	}
}

func startDailyCleanupScheduler(ctx context.Context, auditUsecase *usecase.AuditUsecase, userRepo domain.UserRepository, purgeDays int) {
	ticker := time.NewTicker(24 * time.Hour)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if n, err := auditUsecase.Cleanup(ctx); err != nil {
					slog.ErrorContext(ctx, "audit log cleanup failed", slog.Any("error", err))
				} else {
					slog.Info("audit log cleanup completed", slog.Int64("deleted", n))
				}

				before := time.Now().AddDate(0, 0, -purgeDays)
				users, err := userRepo.ListPendingPurge(ctx, before)
				if err != nil {
					slog.ErrorContext(ctx, "list pending purge users failed", slog.Any("error", err))
					continue
				}
				for _, u := range users {
					if err := userRepo.HardDelete(ctx, u.ID); err != nil {
						slog.ErrorContext(ctx, "hard delete user failed", slog.String("user_id", u.ID), slog.Any("error", err))
					}
				}
				slog.Info("account purge completed", slog.Int("purged", len(users)))
			}
		}
	}()
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("validate config: %w", err)
	}

	setupLogger(cfg.Log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := initApp(ctx, cfg)
	if err != nil {
		return err
	}
	defer a.cleanup()

	startDailyCleanupScheduler(ctx, a.auditUsecase, a.userRepo, cfg.Auth.AccountPurgeDays)

	srv := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:           a.router,
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
