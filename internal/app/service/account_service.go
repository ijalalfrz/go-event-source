package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ijalalfrz/go-event-source/internal/app/dto"
	"github.com/ijalalfrz/go-event-source/internal/app/model"
	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
)

type AccountRepository interface {
	UpsertTx(ctx context.Context, tx *sql.Tx, account *model.Account) error
	WithTransaction(ctx context.Context, txFunc func(context.Context, *sql.Tx) error) error
	FindByID(ctx context.Context, id int64) (model.Account, error)
	FindByIDForUpdateTx(ctx context.Context, tx *sql.Tx, id int64) (model.Account, error)
}

type EventRepository interface {
	CreateTx(ctx context.Context, tx *sql.Tx, event *model.Event) error
	CreateBulkTx(ctx context.Context, tx *sql.Tx, events []model.Event) error
	FindAllByTransactionID(ctx context.Context, transactionID string) ([]model.Event, error)
	FindLastByAggregateID(ctx context.Context, aggregateID int64) (model.Event, error)
}

type AccountService struct {
	accountRepository    AccountRepository
	eventRepository      EventRepository
	requestTimeThreshold time.Duration
	eventVersion         string
}

func NewAccountService(accountRepository AccountRepository,
	eventRepository EventRepository, requestTimeThreshold time.Duration,
	eventVersion string,
) *AccountService {
	return &AccountService{
		accountRepository:    accountRepository,
		eventRepository:      eventRepository,
		requestTimeThreshold: requestTimeThreshold,
		eventVersion:         eventVersion,
	}
}

// CreateAccount godoc
// @Summary      Create Account
// @Description  Create an Account
// @Tags         Account
// @ID           createAccount
// @Produce      json
// @Param        body	body		dto.CreateAccountRequest	true	"Account"
// @Param        x-timestamp	header		string	true	"Request timestamp"
// @Param        x-transaction-id	header		string	true	"Transaction ID"
// @Success      201  "Created"
// @Failure      400  {object}  dto.ErrorResponse	"Bad Request"
// @Failure      404  {object}  dto.ErrorResponse	"Record not found"
// @Failure      409  {object}  dto.ErrorResponse	"Conflict"
// @Failure      500  {object}  dto.ErrorResponse	"Internal Server Error"
// @Router       /accounts [post].
func (s *AccountService) CreateAccount(ctx context.Context, req dto.CreateAccountRequest) error {
	reqContext, err := getRequestContext(ctx, s.requestTimeThreshold)
	if err != nil {
		return fmt.Errorf("failed to get request context: %w", err)
	}

	_, err = s.accountRepository.FindByID(ctx, req.AccountID)
	if err != nil && !errors.Is(err, exception.ErrRecordNotFound) {
		return fmt.Errorf("failed to find account: %w", err)
	}

	if err == nil {
		return ErrAccountAlreadyExists
	}

	// if event already exists, return error for idempotency
	_, err = s.eventRepository.FindAllByTransactionID(ctx, reqContext.TransactionID)
	if err != nil && !errors.Is(err, exception.ErrRecordNotFound) {
		return fmt.Errorf("failed to find events: %w", err)
	}

	if err == nil {
		return ErrIdempotency
	}

	// set event collector
	eventCollector, err := NewAccountEventCollector(ctx, s.eventRepository,
		req.AccountID, reqContext.TransactionID, s.eventVersion)
	if err != nil {
		return fmt.Errorf("failed to create account event collector: %w", err)
	}

	account := &model.Account{
		ID:        req.AccountID,
		Balance:   req.InitialBalance,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// create account within transaction
	err = s.accountRepository.WithTransaction(ctx, func(ctx context.Context, dbTx *sql.Tx) error {
		eventCollector.OnInitBalanceEvent(req.InitialBalance)
		eventCollector.OnDepositReceivedEvent("SYSTEM", req.InitialBalance)

		if err := eventCollector.Place(ctx, dbTx); err != nil {
			return fmt.Errorf("failed to place events: %w", err)
		}

		if err := s.accountRepository.UpsertTx(ctx, dbTx, account); err != nil {
			return fmt.Errorf("failed to upsert account: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create account: %w", err)
	}

	return nil
}

// GetAccount godoc
// @Summary      Get Account
// @Description  Get an Account by ID
// @Tags         Account
// @ID           getAccount
// @Produce      json
// @Param        id	path		string	true	"Account ID"
// @Success      200  {object}  dto.AccountResponse	"Account"
// @Failure      400  {object}  dto.ErrorResponse	"Bad Request"
// @Failure      404  {object}  dto.ErrorResponse	"Record not found"
// @Failure      500  {object}  dto.ErrorResponse	"Internal Server Error"
// @Router       /accounts/{id} [get].
func (s *AccountService) GetAccount(ctx context.Context, req dto.GetAccountRequest) (dto.AccountResponse, error) {
	account, err := s.accountRepository.FindByID(ctx, req.ID)
	if err != nil {
		return dto.AccountResponse{}, fmt.Errorf("failed to get account: %w", err)
	}

	return dto.AccountResponse{
		AccountID: account.ID,
		Balance:   account.Balance,
	}, nil
}
