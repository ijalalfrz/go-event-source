package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/ijalalfrz/go-event-source/internal/app/model"
	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/shopspring/decimal"
)

type AccountEventCollector struct {
	eventRepository EventRepository
	sequenceNumber  int64
	eventVersion    string
	aggregateID     int64
	aggregateType   model.AggregateType
	transactionID   string
	events          []model.Event
}

func NewAccountEventCollector(ctx context.Context, eventRepository EventRepository, aggregateID int64,
	transactionID string, eventVersion string,
) (*AccountEventCollector, error) {
	// get last sequence
	event, err := eventRepository.FindLastByAggregateID(ctx, aggregateID)
	if err != nil && !errors.Is(err, exception.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to find last event: %w", err)
	}

	sequenceNumber := int64(0)
	if err == nil {
		sequenceNumber = event.SequenceNumber
	}

	return &AccountEventCollector{
		eventRepository: eventRepository,
		aggregateID:     aggregateID,
		aggregateType:   model.AggregateTypeAccount,
		transactionID:   transactionID,
		sequenceNumber:  sequenceNumber,
		eventVersion:    eventVersion,
	}, nil
}

func (e *AccountEventCollector) OnInitBalanceEvent(amount decimal.Decimal) {
	payload := map[string]interface{}{
		"initial_balance": amount,
	}
	event := model.Event{
		Version:   e.eventVersion,
		EventType: model.EventTypeInitBalance,
		EventData: payload,
	}
	e.apply(event)
}

func (e *AccountEventCollector) OnDepositReceivedEvent(source string, amount decimal.Decimal) {
	payload := map[string]interface{}{
		"source": source,
		"amount": amount,
	}
	event := model.Event{
		Version:   e.eventVersion,
		EventType: model.EventTypeDepositReceived,
		EventData: payload,
	}
	e.apply(event)
}

func (e *AccountEventCollector) OnSubBalanceEvent(destinationAccountID int64, amount decimal.Decimal) {
	payload := map[string]interface{}{
		"destination_account_id": destinationAccountID,
		"amount":                 amount,
	}

	event := model.Event{
		Version:   e.eventVersion,
		EventType: model.EventTypeDebitBalance,
		EventData: payload,
	}
	e.apply(event)
}

func (e *AccountEventCollector) OnAddBalanceEvent(sourceAccountID int64, amount decimal.Decimal) {
	payload := map[string]interface{}{
		"source_account_id": sourceAccountID,
		"amount":            amount,
	}

	event := model.Event{
		Version:   e.eventVersion,
		EventType: model.EventTypeCreditBalance,
		EventData: payload,
	}
	e.apply(event)
}

func (e *AccountEventCollector) apply(event model.Event) {
	e.sequenceNumber++
	event.SequenceNumber = e.sequenceNumber
	event.AggregateID = e.aggregateID
	event.AggregateType = e.aggregateType
	event.TransactionID = e.transactionID

	e.events = append(e.events, event)
}

func (e *AccountEventCollector) Place(ctx context.Context, tx *sql.Tx) error {
	err := e.eventRepository.CreateBulkTx(ctx, tx, e.events)
	if err != nil {
		return fmt.Errorf("failed to create bulk events: %w", err)
	}

	// reset sequence number and events
	e.sequenceNumber = 0
	e.events = []model.Event{}

	return nil
}
