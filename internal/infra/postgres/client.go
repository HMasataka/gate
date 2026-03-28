package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/HMasataka/gate/internal/domain"
)

type ClientRepo struct {
	db *sqlx.DB
}

func NewClientRepo(db *sqlx.DB) *ClientRepo {
	return &ClientRepo{db: db}
}

func (repo *ClientRepo) ext(ctx context.Context) dbExt {
	return extFromCtx(ctx, repo.db)
}

type clientRow struct {
	ID              string         `db:"id"`
	Secret          string         `db:"secret"`
	Name            string         `db:"name"`
	Type            string         `db:"type"`
	OwnerID         sql.NullString `db:"owner_id"`
	RedirectURIs    pq.StringArray `db:"redirect_uris"`
	AllowedScopes   pq.StringArray `db:"allowed_scopes"`
	TokensRevokedAt sql.NullTime   `db:"tokens_revoked_at"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
}

func (r *clientRow) toDomain() *domain.OAuthClient {
	c := &domain.OAuthClient{
		ID:            r.ID,
		Secret:        r.Secret,
		Name:          r.Name,
		Type:          domain.ClientType(r.Type),
		OwnerID:       r.OwnerID.String,
		RedirectURIs:  []string(r.RedirectURIs),
		AllowedScopes: []string(r.AllowedScopes),
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}

	if r.TokensRevokedAt.Valid {
		t := r.TokensRevokedAt.Time
		c.TokensRevokedAt = &t
	}

	return c
}

func (repo *ClientRepo) Create(ctx context.Context, client *domain.OAuthClient) error {
	const q = `
INSERT INTO oauth_clients (
	id, secret, name, type, owner_id,
	redirect_uris, allowed_scopes,
	tokens_revoked_at
) VALUES (
	:id, :secret, :name, :type, :owner_id,
	:redirect_uris, :allowed_scopes,
	:tokens_revoked_at
)
RETURNING id, created_at, updated_at`

	row := &clientRow{
		ID:            client.ID,
		Secret:        client.Secret,
		Name:          client.Name,
		Type:          string(client.Type),
		OwnerID:       sql.NullString{String: client.OwnerID, Valid: client.OwnerID != ""},
		RedirectURIs:  pq.StringArray(append([]string{}, client.RedirectURIs...)),
		AllowedScopes: pq.StringArray(append([]string{}, client.AllowedScopes...)),
	}

	if client.TokensRevokedAt != nil {
		row.TokensRevokedAt = sql.NullTime{Time: *client.TokensRevokedAt, Valid: true}
	}

	stmt, err := repo.ext(ctx).PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare create client: %w", err)
	}
	defer stmt.Close()

	result := struct {
		ID        string    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&result); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateEmail
		}
		return fmt.Errorf("create client: %w", err)
	}

	client.ID = result.ID
	client.CreatedAt = result.CreatedAt
	client.UpdatedAt = result.UpdatedAt

	return nil
}

func (repo *ClientRepo) GetByID(ctx context.Context, id string) (*domain.OAuthClient, error) {
	const q = `
SELECT id, secret, name, type, owner_id,
	redirect_uris, allowed_scopes,
	tokens_revoked_at, created_at, updated_at
FROM oauth_clients
WHERE id = $1`

	var row clientRow
	if err := repo.ext(ctx).GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get client by id: %w", err)
	}

	return row.toDomain(), nil
}

func (repo *ClientRepo) Update(ctx context.Context, client *domain.OAuthClient) error {
	const q = `
UPDATE oauth_clients SET
	secret           = :secret,
	name             = :name,
	type             = :type,
	owner_id         = :owner_id,
	redirect_uris    = :redirect_uris,
	allowed_scopes   = :allowed_scopes,
	tokens_revoked_at = :tokens_revoked_at,
	updated_at       = NOW()
WHERE id = :id`

	row := &clientRow{
		ID:            client.ID,
		Secret:        client.Secret,
		Name:          client.Name,
		Type:          string(client.Type),
		OwnerID:       sql.NullString{String: client.OwnerID, Valid: client.OwnerID != ""},
		RedirectURIs:  pq.StringArray(append([]string{}, client.RedirectURIs...)),
		AllowedScopes: pq.StringArray(append([]string{}, client.AllowedScopes...)),
	}

	if client.TokensRevokedAt != nil {
		row.TokensRevokedAt = sql.NullTime{Time: *client.TokensRevokedAt, Valid: true}
	}

	stmt, err := repo.ext(ctx).PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare update client: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, row)
	if err != nil {
		return fmt.Errorf("update client: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update client rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *ClientRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM oauth_clients WHERE id = $1`

	result, err := repo.ext(ctx).ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete client: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete client rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *ClientRepo) List(ctx context.Context, offset, limit int) ([]domain.OAuthClient, int, error) {
	const countQ = `SELECT COUNT(*) FROM oauth_clients`

	var total int
	if err := repo.ext(ctx).QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count clients: %w", err)
	}

	const q = `
SELECT id, secret, name, type, owner_id,
	redirect_uris, allowed_scopes,
	tokens_revoked_at, created_at, updated_at
FROM oauth_clients
ORDER BY created_at DESC
LIMIT $1 OFFSET $2`

	var rows []clientRow
	if err := repo.ext(ctx).SelectContext(ctx, &rows, q, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list clients: %w", err)
	}

	clients := make([]domain.OAuthClient, len(rows))
	for i, r := range rows {
		clients[i] = *r.toDomain()
	}

	return clients, total, nil
}

func (repo *ClientRepo) ListByOwner(ctx context.Context, ownerID string, offset, limit int) ([]domain.OAuthClient, int, error) {
	const countQ = `SELECT COUNT(*) FROM oauth_clients WHERE owner_id = $1`

	var total int
	if err := repo.ext(ctx).QueryRowContext(ctx, countQ, ownerID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count clients by owner: %w", err)
	}

	const q = `
SELECT id, secret, name, type, owner_id,
	redirect_uris, allowed_scopes,
	tokens_revoked_at, created_at, updated_at
FROM oauth_clients
WHERE owner_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3`

	var rows []clientRow
	if err := repo.ext(ctx).SelectContext(ctx, &rows, q, ownerID, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list clients by owner: %w", err)
	}

	clients := make([]domain.OAuthClient, len(rows))
	for i, r := range rows {
		clients[i] = *r.toDomain()
	}

	return clients, total, nil
}
