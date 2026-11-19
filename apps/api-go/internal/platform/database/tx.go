package database

import (
	"context"
	"fmt"

	"github.com/orbito/api-go/internal/platform/database/sqlc"
)

// Transactor defines the interface for running operations atomically inside a database transaction.
type Transactor interface {
	WithTx(ctx context.Context, fn func(q *sqlc.Queries) error) error
}

// WithTx runs fn within a database transaction. If fn returns an error or panics,
// the transaction is automatically rolled back. Otherwise, the transaction is committed.
func (db *DB) WithTx(ctx context.Context, fn func(q *sqlc.Queries) error) (err error) {
	if db.pool == nil {
		return fmt.Errorf("database pool is not initialized")
	}

	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		} else if err != nil {
			_ = tx.Rollback(ctx)
		} else {
			if commitErr := tx.Commit(ctx); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	qtx := db.queries.WithTx(tx)
	err = fn(qtx)
	return err
}
