//go:build tools

package tools

import (
	_ "github.com/caarlos0/env/v11"
	_ "github.com/go-chi/chi/v5"
	_ "github.com/go-chi/cors"
	_ "github.com/go-playground/validator/v10"
	_ "github.com/golang-jwt/jwt/v5"
	_ "github.com/golang-migrate/migrate/v4"
	_ "github.com/google/uuid"
	_ "github.com/jackc/pgx/v5"
	_ "github.com/jmoiron/sqlx"
	_ "github.com/prometheus/client_golang/prometheus"
	_ "github.com/pquerna/otp"
	_ "github.com/redis/go-redis/v9"
	_ "golang.org/x/crypto/bcrypt"
)
