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

type RoleRepo struct {
	db *sqlx.DB
}

func NewRoleRepo(db *sqlx.DB) *RoleRepo {
	return &RoleRepo{db: db}
}

type roleRow struct {
	ID          string         `db:"id"`
	Name        string         `db:"name"`
	Description string         `db:"description"`
	ParentID    sql.NullString `db:"parent_id"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
}

func (r *roleRow) toDomain() *domain.Role {
	role := &domain.Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
	if r.ParentID.Valid {
		s := r.ParentID.String
		role.ParentID = &s
	}
	return role
}

func (repo *RoleRepo) Create(ctx context.Context, role *domain.Role) error {
	const q = `
INSERT INTO roles (id, name, description, parent_id)
VALUES (:id, :name, :description, :parent_id)
RETURNING id, created_at, updated_at`

	row := &roleRow{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
	if role.ParentID != nil {
		row.ParentID = sql.NullString{String: *role.ParentID, Valid: true}
	}

	stmt, err := repo.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare create role: %w", err)
	}
	defer stmt.Close()

	result := struct {
		ID        string    `db:"id"`
		CreatedAt time.Time `db:"created_at"`
		UpdatedAt time.Time `db:"updated_at"`
	}{}

	if err := stmt.QueryRowxContext(ctx, row).StructScan(&result); err != nil {
		return fmt.Errorf("create role: %w", err)
	}

	role.ID = result.ID
	role.CreatedAt = result.CreatedAt
	role.UpdatedAt = result.UpdatedAt

	return nil
}

func (repo *RoleRepo) GetByID(ctx context.Context, id string) (*domain.Role, error) {
	const q = `
SELECT id, name, description, parent_id, created_at, updated_at
FROM roles
WHERE id = $1`

	var row roleRow
	if err := repo.db.GetContext(ctx, &row, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get role by id: %w", err)
	}

	return row.toDomain(), nil
}

func (repo *RoleRepo) Update(ctx context.Context, role *domain.Role) error {
	const q = `
UPDATE roles SET
	name        = :name,
	description = :description,
	parent_id   = :parent_id,
	updated_at  = NOW()
WHERE id = :id`

	row := &roleRow{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
	}
	if role.ParentID != nil {
		row.ParentID = sql.NullString{String: *role.ParentID, Valid: true}
	}

	stmt, err := repo.db.PrepareNamedContext(ctx, q)
	if err != nil {
		return fmt.Errorf("prepare update role: %w", err)
	}
	defer stmt.Close()

	result, err := stmt.ExecContext(ctx, row)
	if err != nil {
		return fmt.Errorf("update role: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("update role rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *RoleRepo) Delete(ctx context.Context, id string) error {
	const q = `DELETE FROM roles WHERE id = $1`

	result, err := repo.db.ExecContext(ctx, q, id)
	if err != nil {
		return fmt.Errorf("delete role: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete role rows affected: %w", err)
	}
	if n == 0 {
		return domain.ErrNotFound
	}

	return nil
}

func (repo *RoleRepo) List(ctx context.Context, offset, limit int) ([]domain.Role, int, error) {
	const countQ = `SELECT COUNT(*) FROM roles`
	const listQ = `
SELECT id, name, description, parent_id, created_at, updated_at
FROM roles
ORDER BY created_at ASC
LIMIT $1 OFFSET $2`

	var total int
	if err := repo.db.GetContext(ctx, &total, countQ); err != nil {
		return nil, 0, fmt.Errorf("count roles: %w", err)
	}

	var rows []roleRow
	if err := repo.db.SelectContext(ctx, &rows, listQ, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list roles: %w", err)
	}

	roles := make([]domain.Role, len(rows))
	for i, r := range rows {
		roles[i] = *r.toDomain()
	}

	return roles, total, nil
}

func (repo *RoleRepo) AssignToUser(ctx context.Context, userID, roleID string) error {
	const q = `
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING`

	if _, err := repo.db.ExecContext(ctx, q, userID, roleID); err != nil {
		return fmt.Errorf("assign role to user: %w", err)
	}

	return nil
}

func (repo *RoleRepo) RemoveFromUser(ctx context.Context, userID, roleID string) error {
	const q = `DELETE FROM user_roles WHERE user_id = $1 AND role_id = $2`

	if _, err := repo.db.ExecContext(ctx, q, userID, roleID); err != nil {
		return fmt.Errorf("remove role from user: %w", err)
	}

	return nil
}

func (repo *RoleRepo) GetUserRoles(ctx context.Context, userID string) ([]domain.Role, error) {
	const q = `
SELECT r.id, r.name, r.description, r.parent_id, r.created_at, r.updated_at
FROM roles r
JOIN user_roles ur ON r.id = ur.role_id
WHERE ur.user_id = $1
ORDER BY r.created_at ASC`

	var rows []roleRow
	if err := repo.db.SelectContext(ctx, &rows, q, userID); err != nil {
		return nil, fmt.Errorf("get user roles: %w", err)
	}

	roles := make([]domain.Role, len(rows))
	for i, r := range rows {
		roles[i] = *r.toDomain()
	}

	return roles, nil
}

func (repo *RoleRepo) DetectCycle(ctx context.Context, roleID, parentID string) (bool, error) {
	// Walk up from parentID through the ancestor chain.
	// If roleID appears anywhere in that chain, adding parentID as the parent
	// of roleID would create a cycle.
	const q = `
WITH RECURSIVE ancestors AS (
    SELECT id, parent_id FROM roles WHERE id = $1
    UNION ALL
    SELECT r.id, r.parent_id FROM roles r JOIN ancestors a ON r.id = a.parent_id
)
SELECT EXISTS (SELECT 1 FROM ancestors WHERE id = $2)`

	var exists bool
	if err := repo.db.QueryRowContext(ctx, q, parentID, roleID).Scan(&exists); err != nil {
		return false, fmt.Errorf("detect cycle: %w", err)
	}

	return exists, nil
}
