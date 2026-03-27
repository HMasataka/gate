package postgres

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type txCtxKey struct{}

// dbExt is satisfied by both *sqlx.DB and *sqlx.Tx.
type dbExt interface {
	GetContext(ctx context.Context, dest any, query string, args ...any) error
	SelectContext(ctx context.Context, dest any, query string, args ...any) error
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryxContext(ctx context.Context, query string, args ...any) (*sqlx.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareNamedContext(ctx context.Context, query string) (*sqlx.NamedStmt, error)
}

// injectTx stores tx in the context.
func injectTx(ctx context.Context, tx *sqlx.Tx) context.Context {
	return context.WithValue(ctx, txCtxKey{}, tx)
}

// extFromCtx returns the *sqlx.Tx from ctx if one was injected, otherwise falls back to db.
func extFromCtx(ctx context.Context, db *sqlx.DB) dbExt {
	if tx, ok := ctx.Value(txCtxKey{}).(*sqlx.Tx); ok && tx != nil {
		return tx
	}
	return db
}
