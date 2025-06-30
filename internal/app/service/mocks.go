//go:build unit

package service

import (
	"context"
	"database/sql"

	"github.com/ijalalfrz/go-event-source/internal/app/model"
)

type accountRepositoryMock struct {
	errFindByID                  []error
	errFindByIDForUpdateTx       []error
	errUpsertTx                  []error
	findByIDCallCount            int
	findByIDForUpdateTxCallCount int
	upsertTxCallCount            int
	account                      model.Account
}

func (m *accountRepositoryMock) UpsertTx(ctx context.Context, tx *sql.Tx, account *model.Account) error {
	m.upsertTxCallCount++
	return m.errUpsertTx[m.upsertTxCallCount-1]
}

func (m *accountRepositoryMock) FindByID(ctx context.Context, accountID int64) (model.Account, error) {
	m.findByIDCallCount++
	return m.account, m.errFindByID[m.findByIDCallCount-1]
}

func (m *accountRepositoryMock) FindByIDForUpdateTx(ctx context.Context, tx *sql.Tx, accountID int64) (model.Account, error) {
	m.findByIDForUpdateTxCallCount++
	return m.account, m.errFindByIDForUpdateTx[m.findByIDForUpdateTxCallCount-1]
}

func (m *accountRepositoryMock) WithTransaction(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) error) error {
	return fn(ctx, nil)
}

type eventRepositoryMock struct {
	errCreateTx                     []error
	errCreateBulkTx                 []error
	errFindAllByTransactionID       []error
	errFindLastByAggregateID        []error
	createTxCallCount               int
	createBulkTxCallCount           int
	findAllByTransactionIDCallCount int
	findLastByAggregateIDCallCount  int
	events                          []model.Event
}

func (m *eventRepositoryMock) CreateTx(ctx context.Context, tx *sql.Tx, event *model.Event) error {
	m.createTxCallCount++
	return m.errCreateTx[m.createTxCallCount-1]
}

func (m *eventRepositoryMock) CreateBulkTx(ctx context.Context, tx *sql.Tx, events []model.Event) error {
	m.createBulkTxCallCount++
	return m.errCreateBulkTx[m.createBulkTxCallCount-1]
}

func (m *eventRepositoryMock) FindAllByTransactionID(ctx context.Context, transactionID string) ([]model.Event, error) {
	m.findAllByTransactionIDCallCount++
	return m.events, m.errFindAllByTransactionID[m.findAllByTransactionIDCallCount-1]
}

func (m *eventRepositoryMock) FindLastByAggregateID(ctx context.Context, aggregateID int64) (model.Event, error) {
	m.findLastByAggregateIDCallCount++
	return m.events[0], m.errFindLastByAggregateID[m.findLastByAggregateIDCallCount-1]
}
