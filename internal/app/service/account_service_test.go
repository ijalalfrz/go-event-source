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

func TestAccountService_CreateAccount(t *testing.T) {

	testCreateAccount := func(
		req dto.CreateAccountRequest,
		svc *AccountService,
		ctx context.Context,
		wantErr error,
	) func(t *testing.T) {
		return func(t *testing.T) {
			got := svc.CreateAccount(ctx, req)

			if wantErr != nil {
				assert.Error(t, got)
				assert.Contains(t, got.Error(), wantErr.Error())
			} else {
				assert.NoError(t, got)
			}
		}
	}

	// error request context
	t.Run("error_request_context", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		accountRepository: &accountRepositoryMock{},
	}, context.Background(), fmt.Errorf("request context not found")))

	ctx := createContextWithRequestContext(dto.RequestContext{
		Signature:     "test-signature",
		Language:      "en",
		Timestamp:     time.Now(),
		TransactionID: "tx-12345",
	})

	// error idempotency
	t.Run("error_conflict", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{nil},
		},
		eventVersion: "1.0.0",
	}, ctx, ErrAccountAlreadyExists))

	// error find account
	t.Run("error_find_account", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{errors.New("internal db error")},
		},
	}, ctx, errors.New("internal db error")))

	// transaction id exists
	t.Run("error_idempotency", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{exception.ErrRecordNotFound},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{nil},
		},
		eventVersion: "1.0.0",
	}, ctx, ErrIdempotency))

	// error find events
	t.Run("error_find_events", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{exception.ErrRecordNotFound},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{errors.New("internal db error")},
		},
	}, ctx, errors.New("internal db error")))

	// error failed get last event
	t.Run("error_failed_get_last_event", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{exception.ErrRecordNotFound},
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
	}, ctx, errors.New("internal db error")))

	// failed place events
	t.Run("error_failed_place_events", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{exception.ErrRecordNotFound},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{errors.New("internal db error")},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
			},
		},
	}, ctx, errors.New("internal db error")))

	// failed upsert account
	t.Run("error_failed_upsert_account", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{exception.ErrRecordNotFound},
			errUpsertTx: []error{errors.New("internal db error")},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
			},
		},
	}, ctx, errors.New("internal db error")))

	// success
	t.Run("success", testCreateAccount(dto.CreateAccountRequest{
		AccountID:      1,
		InitialBalance: decimal.NewFromInt(1000),
	}, &AccountService{
		requestTimeThreshold: 30 * time.Second,
		accountRepository: &accountRepositoryMock{
			errFindByID: []error{exception.ErrRecordNotFound},
			errUpsertTx: []error{nil},
		},
		eventRepository: &eventRepositoryMock{
			errFindAllByTransactionID: []error{exception.ErrRecordNotFound},
			errFindLastByAggregateID:  []error{exception.ErrRecordNotFound},
			errCreateBulkTx:           []error{nil},
			events: []model.Event{
				{
					AggregateID:    1,
					SequenceNumber: 1,
				},
			},
		},
	}, ctx, nil))
}
