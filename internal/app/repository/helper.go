package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type transactable struct {
	db *sql.DB
}

func (r *transactable) WithTransaction(ctx context.Context,
	txFunc func(context.Context, *sql.Tx) error,
) error {
	var err error

	dbTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			dbTx.Rollback() //nolint:errcheck
			panic(p)
		} else if err != nil {
			dbTx.Rollback() //nolint:errcheck
		} else {
			err = dbTx.Commit()
		}
	}()

	err = txFunc(ctx, dbTx)

	return err
}
