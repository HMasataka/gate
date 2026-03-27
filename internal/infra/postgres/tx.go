package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TxManager implements domain.TxRunner using PostgreSQL transactions.
type TxManager struct {
	db *sqlx.DB
}

// NewTxManager creates a new TxManager.
func NewTxManager(db *sqlx.DB) *TxManager {
	return &TxManager{db: db}
}

// RunInTx begins a transaction, injects it into ctx, and calls fn.
// Rolls back on error or panic; commits on success.
func (m *TxManager) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := m.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()
	txCtx := injectTx(ctx, tx)
	if err := fn(txCtx); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}
