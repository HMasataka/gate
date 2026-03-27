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

type AuthorizationCodeRepo struct {
	db *sqlx.DB
}

func NewAuthorizationCodeRepo(db *sqlx.DB) *AuthorizationCodeRepo {
	return &AuthorizationCodeRepo{db: db}
}

func (repo *AuthorizationCodeRepo) ext(ctx context.Context) dbExt {
	return extFromCtx(ctx, repo.db)
}

type authorizationCodeRow struct {
	ID                  string         `db:"id"`
	Code                string         `db:"code"`
	ClientID            string         `db:"client_id"`
	UserID              string         `db:"user_id"`
	RedirectURI         string         `db:"redirect_uri"`
	Scopes              pq.StringArray `db:"scopes"`
	CodeChallenge       string         `db:"code_challenge"`
	CodeChallengeMethod string         `db:"code_challenge_method"`
	Nonce               string         `db:"nonce"`
	ExpiresAt           time.Time      `db:"expires_at"`
	UsedAt              sql.NullTime   `db:"used_at"`
	CreatedAt           time.Time      `db:"created_at"`
}

func (r *authorizationCodeRow) toDomain() *domain.AuthorizationCode {
	c := &domain.AuthorizationCode{
		ID:                  r.ID,
		Code:                r.Code,
		ClientID:            r.ClientID,
		UserID:              r.UserID,
		RedirectURI:         r.RedirectURI,
		Scopes:              []string(r.Scopes),
		CodeChallenge:       r.CodeChallenge,
		CodeChallengeMethod: r.CodeChallengeMethod,
		Nonce:               r.Nonce,
		ExpiresAt:           r.ExpiresAt,
		CreatedAt:           r.CreatedAt,
	}

	if r.UsedAt.Valid {
		t := r.UsedAt.Time
		c.UsedAt = &t
	}

	return c
}

func (repo *AuthorizationCodeRepo) Create(ctx context.Context, code *domain.AuthorizationCode) error {
	const q = `
INSERT INTO authorization_codes (
	id, code, client_id, user_id, redirect_uri,
	scopes, code_challenge, code_challenge_method, nonce, expires_at
) VALUES (
	:id, :code, :client_id, :user_id, :redirect_uri,
	:scopes, :code_challenge, :code_challenge_method, :nonce, :expires_at
)
RETURNING id, created_at`

	row := &authorizationCodeRow{
		ID:                  code.ID,
		Code:                code.Code,
		ClientID:            code.ClientID,
		UserID:              code.UserID,
		RedirectURI:         code.RedirectURI,
		Scopes:              pq.StringArray(code.Scopes),
		CodeChallenge:       code.CodeChallenge,
		CodeChallengeMethod: code.CodeChallengeMethod,
		Nonce:               code.Nonce,
		ExpiresAt:           code.ExpiresAt,
	}

	stmt, err := repo.ext(ctx).PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare create authorization code: %w", err)
	}
	defer stmt.Close()

	result := struct {
		ID        string    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
	}{}

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&result); err != nil {
		return fmt.Errorf("create authorization code: %w", err)
	}

	code.ID = result.ID
	code.CreatedAt = result.CreatedAt

	return nil
}

func (repo *AuthorizationCodeRepo) GetByCode(ctx context.Context, codeHash string) (*domain.AuthorizationCode, error) {
	const q = `
SELECT id, code, client_id, user_id, redirect_uri,
	scopes, code_challenge, code_challenge_method, nonce,
	expires_at, used_at, created_at
FROM authorization_codes
WHERE code = $1`

	var row authorizationCodeRow
	if err := repo.ext(ctx).GetContext(ctx, &row, q, codeHash); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get authorization code by code: %w", err)
	}

	return row.toDomain(), nil
}

func (repo *AuthorizationCodeRepo) MarkUsed(ctx context.Context, id string) error {
	const q = `
UPDATE authorization_codes
SET used_at = NOW()
WHERE id = $1`

	if _, err := repo.ext(ctx).ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("mark authorization code used: %w", err)
	}

	return nil
}

func (repo *AuthorizationCodeRepo) DeleteExpired(ctx context.Context) (int64, error) {
	const q = `
DELETE FROM authorization_codes
WHERE expires_at < NOW()`

	result, err := repo.ext(ctx).ExecContext(ctx, q)
	if err != nil {
		return 0, fmt.Errorf("delete expired authorization codes: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete expired authorization codes rows affected: %w", err)
	}

	return n, nil
}
