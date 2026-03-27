package domain

import (
	"context"
	"time"
)

type PasswordHasher interface {
	Hash(ctx context.Context, password string) (string, error)
	Compare(ctx context.Context, password, hash string) (bool, error)
}

type JWTManager interface {
	GenerateAccessToken(ctx context.Context, claims map[string]any) (string, error)
	GenerateIDToken(ctx context.Context, claims map[string]any) (string, error)
	ValidateToken(ctx context.Context, tokenString string) (map[string]any, error)
	JWKS() map[string]any
}

type Mailer interface {
	SendEmailVerification(ctx context.Context, email, token string) error
	SendPasswordReset(ctx context.Context, email, token string) error
}

type SessionStore interface {
	Create(ctx context.Context, session *Session) error
	Get(ctx context.Context, id string) (*Session, error)
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
	ListByUserID(ctx context.Context, userID string) ([]Session, error)
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
	Remaining(ctx context.Context, key string, limit int, window time.Duration) (int, time.Duration, error)
}

type PermissionResolver interface {
	Resolve(ctx context.Context, userID string) ([]string, error)
	Invalidate(ctx context.Context, userID string) error
}

type CaptchaVerifier interface {
	Verify(ctx context.Context, token string) (bool, error)
}

type RandomGenerator interface {
	GenerateToken(length int) (string, error)
	GenerateUUID() string
}
