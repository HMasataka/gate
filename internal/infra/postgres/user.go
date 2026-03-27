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

type UserRepo struct {
	db *sqlx.DB
}

func NewUserRepo(db *sqlx.DB) *UserRepo {
	return &UserRepo{db: db}
}

type userRow struct {
	ID                  string         `db:"id"`
	Email               string         `db:"email"`
	PasswordHash        string         `db:"password_hash"`
	Status              string         `db:"status"`
	EmailVerified       bool           `db:"email_verified"`
	EmailVerifyToken    sql.NullString `db:"email_verify_token"`
	EmailVerifyExpiry   sql.NullTime   `db:"email_verify_expiry"`
	PasswordResetToken  sql.NullString `db:"password_reset_token"`
	PasswordResetExpiry sql.NullTime   `db:"password_reset_expiry"`
	TOTPSecret          sql.NullString `db:"totp_secret"`
	TOTPEnabled         bool           `db:"totp_enabled"`
	RecoveryCodes       pq.StringArray `db:"recovery_codes"`
	TokensRevokedAt     sql.NullTime   `db:"tokens_revoked_at"`
	Version             int            `db:"version"`
	CreatedAt           time.Time      `db:"created_at"`
	UpdatedAt           time.Time      `db:"updated_at"`
	DeletedAt           sql.NullTime   `db:"deleted_at"`
}

func (r *userRow) toDomain() *domain.User {
	u := &domain.User{
		ID:            r.ID,
		Email:         r.Email,
		PasswordHash:  r.PasswordHash,
		Status:        domain.UserStatus(r.Status),
		EmailVerified: r.EmailVerified,
		TOTPEnabled:   r.TOTPEnabled,
		Version:       r.Version,
		CreatedAt:     r.CreatedAt,
		UpdatedAt:     r.UpdatedAt,
	}

	if r.EmailVerifyToken.Valid {
		u.EmailVerifyToken = r.EmailVerifyToken.String
	}
	if r.EmailVerifyExpiry.Valid {
		t := r.EmailVerifyExpiry.Time
		u.EmailVerifyExpiry = &t
	}
	if r.PasswordResetToken.Valid {
		u.PasswordResetToken = r.PasswordResetToken.String
	}
	if r.PasswordResetExpiry.Valid {
		t := r.PasswordResetExpiry.Time
		u.PasswordResetExpiry = &t
	}
	if r.TOTPSecret.Valid {
		u.TOTPSecret = r.TOTPSecret.String
	}
	if r.TokensRevokedAt.Valid {
		t := r.TokensRevokedAt.Time
		u.TokensRevokedAt = &t
	}
	if r.DeletedAt.Valid {
		t := r.DeletedAt.Time
		u.DeletedAt = &t
	}

	u.RecoveryCodes = []string(r.RecoveryCodes)

	return u
}

func (repo *UserRepo) Create(ctx context.Context, user *domain.User) error {
	const q = `
INSERT INTO users (
	id, email, password_hash, status, email_verified,
	email_verify_token, email_verify_expiry,
	password_reset_token, password_reset_expiry,
	totp_secret, totp_enabled, recovery_codes,
	tokens_revoked_at, version
) VALUES (
	:id, :email, :password_hash, :status, :email_verified,
	:email_verify_token, :email_verify_expiry,
	:password_reset_token, :password_reset_expiry,
	:totp_secret, :totp_enabled, :recovery_codes,
	:tokens_revoked_at, :version
)
RETURNING id, created_at, updated_at`

	row := &userRow{
		ID:            user.ID,
		Email:         user.Email,
		PasswordHash:  user.PasswordHash,
		Status:        string(user.Status),
		EmailVerified: user.EmailVerified,
		TOTPEnabled:   user.TOTPEnabled,
		RecoveryCodes: pq.StringArray(user.RecoveryCodes),
		Version:       user.Version,
	}

	if user.EmailVerifyToken != "" {
		row.EmailVerifyToken = sql.NullString{String: user.EmailVerifyToken, Valid: true}
	}
	if user.EmailVerifyExpiry != nil {
		row.EmailVerifyExpiry = sql.NullTime{Time: *user.EmailVerifyExpiry, Valid: true}
	}
	if user.PasswordResetToken != "" {
		row.PasswordResetToken = sql.NullString{String: user.PasswordResetToken, Valid: true}
	}
	if user.PasswordResetExpiry != nil {
		row.PasswordResetExpiry = sql.NullTime{Time: *user.PasswordResetExpiry, Valid: true}
	}
	if user.TOTPSecret != "" {
		row.TOTPSecret = sql.NullString{String: user.TOTPSecret, Valid: true}
	}
	if user.TokensRevokedAt != nil {
		row.TokensRevokedAt = sql.NullTime{Time: *user.TokensRevokedAt, Valid: true}
	}

	stmt, err := repo.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare create user: %w", err)
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
		return fmt.Errorf("create user: %w", err)
	}

	user.ID = result.ID
	user.CreatedAt = result.CreatedAt
	user.UpdatedAt = result.UpdatedAt

	return nil
}

func (repo *UserRepo) GetByID(ctx context.Context, id string) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, status, email_verified,
	email_verify_token, email_verify_expiry,
	password_reset_token, password_reset_expiry,
	totp_secret, totp_enabled, recovery_codes,
	tokens_revoked_at, version, created_at, updated_at, deleted_at
FROM users
WHERE id = $1 AND deleted_at IS NULL`

	var row userRow
	if err := repo.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return row.toDomain(), nil
}

func (repo *UserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	const q = `
SELECT id, email, password_hash, status, email_verified,
	email_verify_token, email_verify_expiry,
	password_reset_token, password_reset_expiry,
	totp_secret, totp_enabled, recovery_codes,
	tokens_revoked_at, version, created_at, updated_at, deleted_at
FROM users
WHERE email = $1 AND deleted_at IS NULL`

	var row userRow
	if err := repo.db.GetContext(ctx, &row, q, email); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return row.toDomain(), nil
}

func (repo *UserRepo) Update(ctx context.Context, user *domain.User) error {
	const q = `
UPDATE users SET
	email               = :email,
	password_hash       = :password_hash,
	status              = :status,
	email_verified      = :email_verified,
	email_verify_token  = :email_verify_token,
	email_verify_expiry = :email_verify_expiry,
	password_reset_token  = :password_reset_token,
	password_reset_expiry = :password_reset_expiry,
	totp_secret         = :totp_secret,
	totp_enabled        = :totp_enabled,
	recovery_codes      = :recovery_codes,
	tokens_revoked_at   = :tokens_revoked_at,
	version             = version + 1,
	updated_at          = NOW()
WHERE id = :id AND version = :version AND deleted_at IS NULL`

	row := &userRow{
		ID:            user.ID,
		Email:         user.Email,
		PasswordHash:  user.PasswordHash,
		Status:        string(user.Status),
		EmailVerified: user.EmailVerified,
		TOTPEnabled:   user.TOTPEnabled,
		RecoveryCodes: pq.StringArray(user.RecoveryCodes),
		Version:       user.Version,
	}

	if user.EmailVerifyToken != "" {
		row.EmailVerifyToken = sql.NullString{String: user.EmailVerifyToken, Valid: true}
	}
	if user.EmailVerifyExpiry != nil {
		row.EmailVerifyExpiry = sql.NullTime{Time: *user.EmailVerifyExpiry, Valid: true}
	}
	if user.PasswordResetToken != "" {
		row.PasswordResetToken = sql.NullString{String: user.PasswordResetToken, Valid: true}
	}
	if user.PasswordResetExpiry != nil {
		row.PasswordResetExpiry = sql.NullTime{Time: *user.PasswordResetExpiry, Valid: true}
	}
	if user.TOTPSecret != "" {
		row.TOTPSecret = sql.NullString{String: user.TOTPSecret, Valid: true}
	}
	if user.TokensRevokedAt != nil {
		row.TokensRevokedAt = sql.NullTime{Time: *user.TokensRevokedAt, Valid: true}
	}

	stmt, err := repo.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare update user: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, row)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrOptimisticLock
	}

	user.Version++

	return nil
}

func (repo *UserRepo) SoftDelete(ctx context.Context, id string) error {
	const q = `
UPDATE users
SET status = 'deleted', deleted_at = NOW(), updated_at = NOW()
WHERE id = $1 AND deleted_at IS NULL`

	result, err := repo.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("soft delete user rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *UserRepo) ListPendingPurge(ctx context.Context, before time.Time) ([]domain.User, error) {
	const q = `
SELECT id, email, password_hash, status, email_verified,
	email_verify_token, email_verify_expiry,
	password_reset_token, password_reset_expiry,
	totp_secret, totp_enabled, recovery_codes,
	tokens_revoked_at, version, created_at, updated_at, deleted_at
FROM users
WHERE status = 'deleted' AND deleted_at < $1`

	var rows []userRow
	if err := repo.db.SelectContext(ctx, &rows, q, before); err != nil {
		return nil, fmt.Errorf("list pending purge users: %w", err)
	}

	users := make([]domain.User, len(rows))
	for i, r := range rows {
		users[i] = *r.toDomain()
	}

	return users, nil
}

func (repo *UserRepo) HardDelete(ctx context.Context, id string) error {
	const q = `DELETE FROM users WHERE id = $1`

	if _, err := repo.db.ExecContext(ctx, q, id); err != nil {
		return fmt.Errorf("hard delete user: %w", err)
	}

	return nil
}
