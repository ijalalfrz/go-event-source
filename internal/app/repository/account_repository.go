package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ijalalfrz/go-event-source/internal/app/model"
	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
)

type AccountRepository struct {
	db *sql.DB
	transactable
	errorMapper
}

func NewAccountRepository(db *sql.DB) *AccountRepository {
	return &AccountRepository{
		db:           db,
		transactable: transactable{db: db},
	}
}

func (r *AccountRepository) UpsertTx(ctx context.Context, dbTx *sql.Tx, account *model.Account) error {
	if dbTx == nil {
		return errors.New("transaction is nil")
	}

	query := `
		INSERT INTO accounts (id, balance, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET balance = $2, updated_at = $4
	`

	stmt, err := dbTx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, account.ID, account.Balance, account.CreatedAt, account.UpdatedAt)
	if err != nil {
		err = r.mapError(err)

		return fmt.Errorf("failed to exec statement: %w", err)
	}

	return nil
}

func (r *AccountRepository) FindByID(ctx context.Context, accountID int64) (model.Account, error) {
	query := `
		SELECT id, balance
		FROM accounts
		WHERE id = $1
	`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return model.Account{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, accountID)

	var account model.Account

	err = row.Scan(&account.ID, &account.Balance)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err := exception.ErrRecordNotFound
		err.MessageVars = map[string]interface{}{
			"name": "account",
		}

		return model.Account{}, fmt.Errorf("account not found: %w", err)
	}

	if err != nil {
		err = r.mapError(err)

		return model.Account{}, fmt.Errorf("failed to scan row: %w", err)
	}

	return account, nil
}

func (r *AccountRepository) FindByIDForUpdateTx(ctx context.Context,
	dbTx *sql.Tx, accountID int64,
) (model.Account, error) {
	query := `
		SELECT id, balance
		FROM accounts
		WHERE id = $1
		FOR UPDATE
	`

	stmt, err := dbTx.PrepareContext(ctx, query)
	if err != nil {
		return model.Account{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, accountID)

	var account model.Account

	err = row.Scan(&account.ID, &account.Balance)
	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err := exception.ErrRecordNotFound
		err.MessageVars = map[string]interface{}{
			"name": "account",
		}

		return model.Account{}, fmt.Errorf("account not found: %w", err)
	}

	if err != nil {
		err = r.mapError(err)

		return model.Account{}, fmt.Errorf("failed to scan row: %w", err)
	}

	return account, nil
}
