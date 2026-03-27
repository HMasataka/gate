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

type PermissionRepo struct {
	db *sqlx.DB
}

func NewPermissionRepo(db *sqlx.DB) *PermissionRepo {
	return &PermissionRepo{db: db}
}

type permissionRow struct {
	ID          string    `db:"id"`
	Name        string    `db:"name"`
	Description string    `db:"description"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}

func (r *permissionRow) toDomain() *domain.Permission {
	return &domain.Permission{
		ID:          r.ID,
		Name:        r.Name,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}

func (r *PermissionRepo) Create(ctx context.Context, perm *domain.Permission) error {
	query := `INSERT INTO permissions (id, name, description, created_at, updated_at)
	          VALUES (:id, :name, :description, :created_at, :updated_at)`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare create permission: %w", err)
	}
	defer stmt.Close()

	row := permissionRow{
		ID:          perm.ID,
		Name:        perm.Name,
		Description: perm.Description,
		CreatedAt:   perm.CreatedAt,
		UpdatedAt:   perm.UpdatedAt,
	}
	if _, err := stmt.ExecContext(ctx, row); err != nil {
		return fmt.Errorf("create permission: %w", err)
	}
	return nil
}

func (r *PermissionRepo) GetByID(ctx context.Context, id string) (*domain.Permission, error) {
	var row permissionRow
	if err := r.db.GetContext(ctx, &row, `SELECT id, name, description, created_at, updated_at FROM permissions WHERE id = $1`, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("get permission: %w", err)
	}
	return row.toDomain(), nil
}

func (r *PermissionRepo) Update(ctx context.Context, perm *domain.Permission) error {
	query := `UPDATE permissions SET name = :name, description = :description, updated_at = :updated_at WHERE id = :id`
	stmt, err := r.db.PrepareNamedContext(ctx, query)
	if err != nil {
		return fmt.Errorf("prepare update permission: %w", err)
	}
	defer stmt.Close()

	row := permissionRow{
		ID:          perm.ID,
		Name:        perm.Name,
		Description: perm.Description,
		UpdatedAt:   perm.UpdatedAt,
	}
	res, err := stmt.ExecContext(ctx, row)
	if err != nil {
		return fmt.Errorf("update permission: %w", err)
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *PermissionRepo) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM permissions WHERE id = $1`, id); err != nil {
		return fmt.Errorf("delete permission: %w", err)
	}
	return nil
}

func (r *PermissionRepo) List(ctx context.Context, offset, limit int) ([]domain.Permission, int, error) {
	var total int
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(*) FROM permissions`); err != nil {
		return nil, 0, fmt.Errorf("count permissions: %w", err)
	}

	var rows []permissionRow
	if err := r.db.SelectContext(ctx, &rows, `SELECT id, name, description, created_at, updated_at FROM permissions ORDER BY created_at LIMIT $1 OFFSET $2`, limit, offset); err != nil {
		return nil, 0, fmt.Errorf("list permissions: %w", err)
	}

	perms := make([]domain.Permission, len(rows))
	for i, row := range rows {
		perms[i] = *row.toDomain()
	}
	return perms, total, nil
}

func (r *PermissionRepo) AssignToRole(ctx context.Context, roleID, permID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, roleID, permID)
	if err != nil {
		return fmt.Errorf("assign permission to role: %w", err)
	}
	return nil
}

func (r *PermissionRepo) RemoveFromRole(ctx context.Context, roleID, permID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM role_permissions WHERE role_id = $1 AND permission_id = $2`, roleID, permID)
	if err != nil {
		return fmt.Errorf("remove permission from role: %w", err)
	}
	return nil
}

func (r *PermissionRepo) AssignToUser(ctx context.Context, userID, permID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO user_permissions (user_id, permission_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`, userID, permID)
	if err != nil {
		return fmt.Errorf("assign permission to user: %w", err)
	}
	return nil
}

func (r *PermissionRepo) RemoveFromUser(ctx context.Context, userID, permID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM user_permissions WHERE user_id = $1 AND permission_id = $2`, userID, permID)
	if err != nil {
		return fmt.Errorf("remove permission from user: %w", err)
	}
	return nil
}

// ResolveForUser returns all permission names for a user (direct + via roles, recursively).
func (r *PermissionRepo) ResolveForUser(ctx context.Context, userID string) ([]string, error) {
	query := `
WITH RECURSIVE role_tree AS (
    SELECT r.id FROM roles r
    JOIN user_roles ur ON ur.role_id = r.id WHERE ur.user_id = $1
    UNION
    SELECT r.id FROM roles r JOIN role_tree rt ON rt.id = r.parent_id
)
SELECT DISTINCT p.name FROM permissions p
WHERE p.id IN (
    SELECT rp.permission_id FROM role_permissions rp JOIN role_tree rt ON rp.role_id = rt.id
    UNION
    SELECT up.permission_id FROM user_permissions up WHERE up.user_id = $1
)`
	var names []string
	if err := r.db.SelectContext(ctx, &names, query, userID); err != nil {
		return nil, fmt.Errorf("resolve permissions for user: %w", err)
	}
	return names, nil
}
