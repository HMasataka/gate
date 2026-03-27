package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/HMasataka/gate/internal/domain"
)

// SocialConnectionRepo implements domain.SocialConnectionRepository using PostgreSQL.
type SocialConnectionRepo struct {
	db *sqlx.DB
}

// NewSocialConnectionRepo creates a new SocialConnectionRepo.
func NewSocialConnectionRepo(db *sqlx.DB) *SocialConnectionRepo {
	return &SocialConnectionRepo{db: db}
}

func (repo *SocialConnectionRepo) ext(ctx context.Context) dbExt {
	return extFromCtx(ctx, repo.db)
}

type socialConnectionRow struct {
	ID             string    `db:"id"`
	UserID         string    `db:"user_id"`
	Provider       string    `db:"provider"`
	ProviderUserID string    `db:"provider_user_id"`
	Email          string    `db:"email"`
	Name           string    `db:"name"`
	AvatarURL      string    `db:"avatar_url"`
	CreatedAt      time.Time `db:"created_at"`
	UpdatedAt      time.Time `db:"updated_at"`
}

func (r *socialConnectionRow) toDomain() *domain.SocialConnection {
	return &domain.SocialConnection{
		ID:             r.ID,
		UserID:         r.UserID,
		Provider:       r.Provider,
		ProviderUserID: r.ProviderUserID,
		Email:          r.Email,
		Name:           r.Name,
		AvatarURL:      r.AvatarURL,
		CreatedAt:      r.CreatedAt,
		UpdatedAt:      r.UpdatedAt,
	}
}

// Create inserts a new social connection record.
func (repo *SocialConnectionRepo) Create(ctx context.Context, conn *domain.SocialConnection) error {
	const q = `
INSERT INTO social_connections (
	id, user_id, provider, provider_user_id, email, name, avatar_url
) VALUES (
	:id, :user_id, :provider, :provider_user_id, :email, :name, :avatar_url
)
RETURNING id, created_at, updated_at`

	row := &socialConnectionRow{
		ID:             conn.ID,
		UserID:         conn.UserID,
		Provider:       conn.Provider,
		ProviderUserID: conn.ProviderUserID,
		Email:          conn.Email,
		Name:           conn.Name,
		AvatarURL:      conn.AvatarURL,
	}

	stmt, err := repo.ext(ctx).PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare create social connection: %w", err)
	}
	defer stmt.Close()

	result := struct {
		ID        string    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&result); err != nil {
		return fmt.Errorf("create social connection: %w", err)
	}

	conn.ID = result.ID
	conn.CreatedAt = result.CreatedAt
	conn.UpdatedAt = result.UpdatedAt

	return nil
}

// GetByProviderAndProviderUserID retrieves a social connection by provider and provider user ID.
func (repo *SocialConnectionRepo) GetByProviderAndProviderUserID(ctx context.Context, provider, providerUserID string) (*domain.SocialConnection, error) {
	const q = `
SELECT id, user_id, provider, provider_user_id, email, name, avatar_url, created_at, updated_at
FROM social_connections
WHERE provider = $1 AND provider_user_id = $2`

	var row socialConnectionRow
	if err := repo.ext(ctx).GetContext(ctx, &row, q, provider, providerUserID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get social connection by provider and provider_user_id: %w", err)
	}

	return row.toDomain(), nil
}

// GetByUserID retrieves all social connections for a user.
func (repo *SocialConnectionRepo) GetByUserID(ctx context.Context, userID string) ([]domain.SocialConnection, error) {
	const q = `
SELECT id, user_id, provider, provider_user_id, email, name, avatar_url, created_at, updated_at
FROM social_connections
WHERE user_id = $1`

	var rows []socialConnectionRow
	if err := repo.ext(ctx).SelectContext(ctx, &rows, q, userID); err != nil {
		return nil, fmt.Errorf("get social connections by user_id: %w", err)
	}

	conns := make([]domain.SocialConnection, len(rows))
	for i, r := range rows {
		conns[i] = *r.toDomain()
	}

	return conns, nil
}

// Update updates an existing social connection record.
func (repo *SocialConnectionRepo) Update(ctx context.Context, conn *domain.SocialConnection) error {
	const q = `
UPDATE social_connections SET
	email      = :email,
	name       = :name,
	avatar_url = :avatar_url,
	updated_at = NOW()
WHERE id = :id`

	row := &socialConnectionRow{
		ID:        conn.ID,
		Email:     conn.Email,
		Name:      conn.Name,
		AvatarURL: conn.AvatarURL,
	}

	stmt, err := repo.ext(ctx).PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare update social connection: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, row)
	if err != nil {
		return fmt.Errorf("update social connection: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update social connection rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete removes a social connection by ID.
func (repo *SocialConnectionRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM social_connections WHERE id = $1`

	result, err := repo.ext(ctx).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete social connection: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete social connection rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}
