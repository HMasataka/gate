package postgres

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"

	"github.com/HMasataka/gate/internal/config"
)

func NewDB(ctx context.Context, cfg config.DatabaseConfig) (*sqlx.DB, error) {
	connConfig, err := pgx.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("parse database url: %w", err)
	}

	db := stdlib.OpenDB(*connConfig)
	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.ConnMaxLifetime)
	db.SetConnMaxIdleTime(cfg.ConnMaxIdleTime)

	sqlxDB := sqlx.NewDb(db, "pgx")

	if err := pingWithRetry(ctx, sqlxDB, cfg.ConnectRetries); err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	return sqlxDB, nil
}

func pingWithRetry(ctx context.Context, db *sqlx.DB, maxRetries int) error {
	var err error
	for i := 0; i <= maxRetries; i++ {
		if err = db.PingContext(ctx); err == nil {
			return nil
		}
		if i < maxRetries {
			backoff := min(time.Duration(1<<uint(i))*time.Second, 30*time.Second)
			slog.Warn("database connection failed, retrying",
				"attempt", i+1, "max_retries", maxRetries, "backoff", backoff, "error", err)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}
	}
	return fmt.Errorf("after %d retries: %w", maxRetries, err)
}
