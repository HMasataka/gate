package domain

import "context"

// TxRunner executes fn within a database transaction.
// The transaction is propagated via the context passed to fn.
type TxRunner interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}
