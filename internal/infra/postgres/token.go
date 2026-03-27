package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/HMasataka/gate/internal/domain"
)

type RefreshTokenRepo struct {
	db *sqlx.DB
}

func NewRefreshTokenRepo(db *sqlx.DB) *RefreshTokenRepo {
	return &RefreshTokenRepo{db: db}
}

func (repo *RefreshTokenRepo) ext(ctx context.Context) dbExt {
	return extFromCtx(ctx, repo.db)
}

type refreshTokenRow struct {
	ID        string         `db:"id"`
	TokenHash string         `db:"token_hash"`
	UserID    string         `db:"user_id"`
	ClientID  string         `db:"client_id"`
	FamilyID  string         `db:"family_id"`
	Scopes    pq.StringArray `db:"scopes"`
	ExpiresAt time.Time      `db:"expires_at"`
	RevokedAt sql.NullTime   `db:"revoked_at"`
	CreatedAt time.Time      `db:"created_at"`
}

func (r *refreshTokenRow) toDomain() *domain.RefreshToken {
	t := &domain.RefreshToken{
		ID:        r.ID,
		TokenHash: r.TokenHash,
		UserID:    r.UserID,
		ClientID:  r.ClientID,
		FamilyID:  r.FamilyID,
		Scopes:    []string(r.Scopes),
		ExpiresAt: r.ExpiresAt,
		CreatedAt: r.CreatedAt,
	}

	if r.RevokedAt.Valid {
		rv := r.RevokedAt.Time
		t.RevokedAt = &rv
	}

	return t
}

func (repo *RefreshTokenRepo) Create(ctx context.Context, token *domain.RefreshToken) error {
	const q = `
INSERT INTO refresh_tokens (
	id, token_hash, user_id, client_id, family_id,
	scopes, expires_at
) VALUES (
	:id, :token_hash, :user_id, :client_id, :family_id,
	:scopes, :expires_at
)
RETURNING id, created_at`

	row := &refreshTokenRow{
		ID:        token.ID,
		TokenHash: token.TokenHash,
		UserID:    token.UserID,
		ClientID:  token.ClientID,
		FamilyID:  token.FamilyID,
		Scopes:    pq.StringArray(token.Scopes),
		ExpiresAt: token.ExpiresAt,
	}

	stmt, err := repo.ext(ctx).PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare create refresh token: %w", err)
	}
	defer stmt.Close()

	result := struct {
		ID        string    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}{}

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&result); err != nil {
		return fmt.Errorf("create refresh token: %w", err)
	}

	token.ID = result.ID
	token.CreatedAt = result.CreatedAt

	return nil
}

func (repo *RefreshTokenRepo) GetByTokenHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	const q = `
SELECT id, token_hash, user_id, client_id, family_id,
	scopes, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE token_hash = $1 AND revoked_at IS NULL AND expires_at > NOW()`

	var row refreshTokenRow
	if err := repo.ext(ctx).GetContext(ctx, &row, q, hash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get refresh token by hash: %w", err)
	}

	return row.toDomain(), nil
}

func (repo *RefreshTokenRepo) RevokeByID(ctx context.Context, id string) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE id = $1`

	if _, err := repo.ext(ctx).ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("revoke refresh token by id: %w", err)
	}

	return nil
}

func (repo *RefreshTokenRepo) RevokeByFamilyID(ctx context.Context, familyID string) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE family_id = $1 AND revoked_at IS NULL`

	if _, err := repo.ext(ctx).ExecContext(ctx, q, familyID); err != nil {
		return fmt.Errorf("revoke refresh tokens by family id: %w", err)
	}

	return nil
}

func (repo *RefreshTokenRepo) RevokeByUserID(ctx context.Context, userID string) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE user_id = $1 AND revoked_at IS NULL`

	if _, err := repo.ext(ctx).ExecContext(ctx, q, userID); err != nil {
		return fmt.Errorf("revoke refresh tokens by user id: %w", err)
	}

	return nil
}

func (repo *RefreshTokenRepo) RevokeByClientID(ctx context.Context, clientID string) error {
	const q = `
UPDATE refresh_tokens
SET revoked_at = NOW()
WHERE client_id = $1 AND revoked_at IS NULL`

	if _, err := repo.ext(ctx).ExecContext(ctx, q, clientID); err != nil {
		return fmt.Errorf("revoke refresh tokens by client id: %w", err)
	}

	return nil
}

func (repo *RefreshTokenRepo) DeleteExpired(ctx context.Context) (int64, error) {
	const q = `
DELETE FROM refresh_tokens
WHERE expires_at < NOW()`

	result, err := repo.ext(ctx).ExecContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete expired refresh tokens rows affected: %w", err)
	}

	return n, nil
}
