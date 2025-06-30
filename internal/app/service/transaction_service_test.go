//go:build unit

package service

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/ijalalfrz/go-event-source/internal/app/dto"
	"github.com/ijalalfrz/go-event-source/internal/app/model"
	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
)

func TestTransactionService_Transfer(t *testing.T) {

	testTransfer := func(
		req dto.CreateTransferRequest,
		svc *TransactionService,
		ctx context.Context,
		wantErr error,
	) func(t *testing.T) {
		return func(t *testing.T) {
			got := svc.Transfer(ctx, req)

			if wantErr != nil {
				assert.Error(t, got)
				assert.Contains(t, got.Error(), wantErr.Error())
			} else {
				assert.NoError(t, got)
			}
		}
	}

	// error request context
	t.Run("error_request_context", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		accountRepository: &accountRepositoryMock{},
		eventRepository:   &eventRepositoryMock{},
	}, context.Background(), fmt.Errorf("request context not found")))

	ctx := createContextWithRequestContext(dto.RequestContext{
		Signature:     "test-signature",
		Language:      "en",
		Timestamp:     time.Now(),
		TransactionID: "tx-12345",
	})

	// error idempotency - transaction already exists
	t.Run("error_idempotency", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{nil}, // No error means events found
			events: []model.Event{
				{
					TransactionID: "tx-12345",
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, ErrIdempotency))

	// error source and destination account same
	t.Run("error_source_and_destination_same", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 1, // Same account
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound},
		},
		eventVersion: "1.0.0",
	}, ctx, ErrSourceAndDestinationAccountSame))

	// error find events for idempotency check
	t.Run("error_find_events", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{errors.New("internal db error")},
		},
		eventVersion: "1.0.0",
	}, ctx, errors.New("internal db error")))

	// error same source and destination account
	t.Run("error_same_source_and_destination_account", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 1,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound},
		},
	}, ctx, ErrSourceAndDestinationAccountSame))

	// error create source account event collector
	t.Run("error_create_source_event_collector", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{errors.New("internal db error")},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, errors.New("internal db error")))

	// error create destination account event collector
	t.Run("error_create_destination_event_collector", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, errors.New("internal db error")},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, errors.New("internal db error")))

	// error place source account events
	t.Run("error_place_source_events", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{errors.New("internal db error"), nil, nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, errors.New("internal db error")))

	// error place destination account events
	t.Run("error_place_destination_events", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil, errors.New("internal db error"), nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, errors.New("internal db error")))

	// error source account not found
	t.Run("error_source_account_not_found", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{exception.ErrRecordNotFound},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil, nil, nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, ErrSourceAccountNotFound))

	// error destination account not found
	t.Run("error_destination_account_not_found", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil, exception.ErrRecordNotFound},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil, nil, nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, ErrDestinationAccountNotFound))

	// error insufficient balance
	t.Run("error_insufficient_balance", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(1000), // More than available balance
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil, nil},
			account: model.Account{
				ID:      1,
				Balance: decimal.NewFromInt(500), // Less than transfer amount
			},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil, nil, nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, ErrInsufficientBalance))

	// error upsert source account
	t.Run("error_upsert_source_account", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil, nil},
			errUpsertTx:            []error{errors.New("internal db error"), nil},
			account: model.Account{
				ID:      1,
				Balance: decimal.NewFromInt(1000),
			},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil, nil, nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, errors.New("internal db error")))

	// error upsert destination account
	t.Run("error_upsert_destination_account", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil, nil},
			errUpsertTx:            []error{nil, errors.New("internal db error")},
			account: model.Account{
				ID:      1,
				Balance: decimal.NewFromInt(1000),
			},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil, nil, nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, errors.New("internal db error")))

	// success
	t.Run("success", testTransfer(dto.CreateTransferRequest{
		SourceAccountID:      1,
		DestinationAccountID: 2,
		Amount:               decimal.NewFromInt(100),
	}, &TransactionService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByIDForUpdateTx: []error{nil, nil},
			errUpsertTx:            []error{nil, nil},
			account: model.Account{
				ID:      1,
				Balance: decimal.NewFromInt(1000),
			},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound, exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil, nil, nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
				{
					AggregateID:    2,
					SequenceNumber: 1,
				},
			},
		},
		eventVersion: "1.0.0",
	}, ctx, nil))
}
