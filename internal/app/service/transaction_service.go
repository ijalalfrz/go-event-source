package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/ijalalfrz/go-event-source/internal/app/dto"
	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
)

type TransactionService struct {
	eventRepository      EventRepository
	accountRepository    AccountRepository
	requestTimeThreshold time.Duration
	eventVersion         string
}

func NewTransactionService(accountRepository AccountRepository,
	eventRepository EventRepository, requestTimeThreshold time.Duration,
	eventVersion string,
) *TransactionService {
	return &TransactionService{
		accountRepository:    accountRepository,
		eventRepository:      eventRepository,
		requestTimeThreshold: requestTimeThreshold,
		eventVersion:         eventVersion,
	}
}

// Transfer godoc
// @Summary      Transfer
// @Description  Transfer between two accounts
// @Tags         Transfer
// @ID           transfer
// @Produce      json
// @Param        req body create transfer	body		dto.CreateTransferRequest	true	"Transfer"
// @Param        x-timestamp	header		string	true	"Request timestamp"
// @Param        x-transaction-id	header		string	true	"Transaction ID"
// @Success      204  "No content"
// @Failure      400  {object}  dto.ErrorResponse	"Bad Request"
// @Failure      404  {object}  dto.ErrorResponse	"Record not found"
// @Failure      409  {object}  dto.ErrorResponse	"Conflict"
// @Failure      500  {object}  dto.ErrorResponse	"Internal Server Error"
// @Router       /transfers [post].
func (s *TransactionService) Transfer(ctx context.Context, req dto.CreateTransferRequest) error {
	// get request context for transaction id and timestamp
	reqContext, err := getRequestContext(ctx, s.requestTimeThreshold)
	if err != nil {
		return fmt.Errorf("failed to get request context: %w", err)
	}

	// if event already exists, return err for idempotency
	_, err = s.eventRepository.FindAllByTransactionID(ctx, reqContext.TransactionID)
	if err != nil && !errors.Is(err, exception.ErrRecordNotFound) {
		return fmt.Errorf("failed to find events: %w", err)
	}

	if err == nil {
		return ErrIdempotency
	}

	// validate source and destination account
	if req.SourceAccountID == req.DestinationAccountID {
		return ErrSourceAndDestinationAccountSame
	}

	// set event collector source and destination account
	sourceAccountEventCollector, err := NewAccountEventCollector(ctx, s.eventRepository, req.SourceAccountID,
		reqContext.TransactionID, s.eventVersion)
	if err != nil {
		return fmt.Errorf("failed to create account event collector: %w", err)
	}

	destinationAccountEventCollector, err := NewAccountEventCollector(ctx, s.eventRepository, req.DestinationAccountID,
		reqContext.TransactionID, s.eventVersion)
	if err != nil {
		return fmt.Errorf("failed to create account event collector: %w", err)
	}

	// process transfer within transaction
	err = s.accountRepository.WithTransaction(ctx, func(ctx context.Context, dbTx *sql.Tx) error {
		// add event
		sourceAccountEventCollector.OnSubBalanceEvent(req.SourceAccountID, req.Amount)
		destinationAccountEventCollector.OnAddBalanceEvent(req.DestinationAccountID, req.Amount)

		if err := sourceAccountEventCollector.Place(ctx, dbTx); err != nil {
			return fmt.Errorf("failed to place events: %w", err)
		}

		if err := destinationAccountEventCollector.Place(ctx, dbTx); err != nil {
			return fmt.Errorf("failed to place events: %w", err)
		}

		// update projection
		err = s.processTransfer(ctx, dbTx, req)
		if err != nil {
			return fmt.Errorf("failed to process transfer: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to process transfer: %w", err)
	}

	return nil
}

func (s *TransactionService) processTransfer(ctx context.Context, dbTx *sql.Tx, req dto.CreateTransferRequest) error {
	// find source account
	sourceAccount, err := s.accountRepository.FindByIDForUpdateTx(ctx, dbTx, req.SourceAccountID)
	if err != nil && errors.Is(err, exception.ErrRecordNotFound) {
		err = ErrSourceAccountNotFound

		return fmt.Errorf("failed to find account: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to find account: %w", err)
	}

	// find destination account
	destinationAccount, err := s.accountRepository.FindByIDForUpdateTx(ctx, dbTx, req.DestinationAccountID)
	if err != nil && errors.Is(err, exception.ErrRecordNotFound) {
		err = ErrDestinationAccountNotFound

		return fmt.Errorf("failed to find account: %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to find account: %w", err)
	}

	// validate balance
	if sourceAccount.Balance.LessThan(req.Amount) {
		err = ErrInsufficientBalance

		return fmt.Errorf("insufficient balance: %w", err)
	}

	// update balance
	sourceAccount.Balance = sourceAccount.Balance.Sub(req.Amount)
	destinationAccount.Balance = destinationAccount.Balance.Add(req.Amount)
	sourceAccount.UpdatedAt = time.Now()
	destinationAccount.UpdatedAt = time.Now()

	if err := s.accountRepository.UpsertTx(ctx, dbTx, &sourceAccount); err != nil {
		return fmt.Errorf("failed to upsert account: %w", err)
	}

	if err := s.accountRepository.UpsertTx(ctx, dbTx, &destinationAccount); err != nil {
		return fmt.Errorf("failed to upsert account: %w", err)
	}

	return nil
}
