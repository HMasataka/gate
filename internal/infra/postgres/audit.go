package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/HMasataka/gate/internal/domain"
)

type AuditLogRepo struct {
	db *sqlx.DB
}

func NewAuditLogRepo(db *sqlx.DB) *AuditLogRepo {
	return &AuditLogRepo{db: db}
}

func (repo *AuditLogRepo) ext(ctx context.Context) dbExt {
	return extFromCtx(ctx, repo.db)
}

type auditLogRow struct {
	ID        string         `db:"id"`
	UserID    sql.NullString `db:"user_id"`
	Action    string         `db:"action"`
	IPAddress string         `db:"ip_address"`
	UserAgent string         `db:"user_agent"`
	Metadata  []byte         `db:"metadata"`
	CreatedAt time.Time      `db:"created_at"`
}

func (r *auditLogRow) toDomain() *domain.AuditLog {
	l := &domain.AuditLog{
		ID:        r.ID,
		Action:    domain.AuditAction(r.Action),
		IPAddress: r.IPAddress,
		UserAgent: r.UserAgent,
		CreatedAt: r.CreatedAt,
	}

	if r.UserID.Valid {
		s := r.UserID.String
		l.UserID = &s
	}

	if len(r.Metadata) > 0 {
		var m map[string]any
		if err := json.Unmarshal(r.Metadata, &m); err == nil {
			l.Metadata = m
		}
	}

	return l
}

func (repo *AuditLogRepo) Create(ctx context.Context, log *domain.AuditLog) error {
	const q = `
INSERT INTO audit_logs (id, user_id, action, ip_address, user_agent, metadata)
VALUES ($1, $2, $3, $4, $5, $6)`

	metaJSON, err := json.Marshal(log.Metadata)
	if err != nil {
		return fmt.Errorf("marshal audit log metadata: %w", err)
	}

	var userID sql.NullString
	if log.UserID != nil {
		userID = sql.NullString{String: *log.UserID, Valid: true}
	}

	if _, err := repo.ext(ctx).ExecContext(ctx, q, log.ID, userID, string(log.Action), log.IPAddress, log.UserAgent, metaJSON); err != nil {
		return fmt.Errorf("create audit log: %w", err)
	}

	return nil
}

func (repo *AuditLogRepo) List(ctx context.Context, filter domain.AuditLogFilter) ([]domain.AuditLog, int, error) {
	args := []any{}
	argIdx := 1
	where := " WHERE 1=1"

	if filter.UserID != nil {
		where += fmt.Sprintf(" AND user_id = $%d", argIdx)
		args = append(args, *filter.UserID)
		argIdx++
	}
	if filter.Action != nil {
		where += fmt.Sprintf(" AND action = $%d", argIdx)
		args = append(args, string(*filter.Action))
		argIdx++
	}
	if filter.From != nil {
		where += fmt.Sprintf(" AND created_at >= $%d", argIdx)
		args = append(args, *filter.From)
		argIdx++
	}
	if filter.To != nil {
		where += fmt.Sprintf(" AND created_at <= $%d", argIdx)
		args = append(args, *filter.To)
		argIdx++
	}

	countQ := "SELECT COUNT(*) FROM audit_logs" + where
	var total int
	if err := repo.ext(ctx).GetContext(ctx, &total, countQ, args...); err != nil {
		return nil, 0, fmt.Errorf("count audit logs: %w", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}

	listQ := "SELECT id, user_id, action, ip_address, user_agent, metadata, created_at FROM audit_logs" +
		where +
		fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
	args = append(args, limit, filter.Offset)

	rows, err := repo.ext(ctx).QueryxContext(ctx, listQ, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []domain.AuditLog{}, total, nil
		}
		return nil, 0, fmt.Errorf("list audit logs: %w", err)
	}
	defer rows.Close()

	var logs []domain.AuditLog
	for rows.Next() {
		var row auditLogRow
		if err := rows.StructScan(&row); err != nil {
			return nil, 0, fmt.Errorf("scan audit log: %w", err)
		}
		logs = append(logs, *row.toDomain())
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate audit logs: %w", err)
	}

	if logs == nil {
		logs = []domain.AuditLog{}
	}

	return logs, total, nil
}

func (repo *AuditLogRepo) DeleteBefore(ctx context.Context, before time.Time) (int64, error) {
	const q = `DELETE FROM audit_logs WHERE created_at < $1`

	result, err := repo.ext(ctx).ExecContext(ctx, q, before)
	if err != nil {
		return 0, fmt.Errorf("delete audit logs before %s: %w", before, err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("delete audit logs rows affected: %w", err)
	}

	return n, nil
}
