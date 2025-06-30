package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ijalalfrz/go-event-source/internal/app/model"
	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/lib/pq"
)

type EventRepository struct {
	db *sql.DB
	transactable
	errorMapper
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{
		db:           db,
		transactable: transactable{db: db},
	}
}

func (r *EventRepository) CreateTx(ctx context.Context, dbTx *sql.Tx, event *model.Event) error {
	if dbTx == nil {
		return errors.New("transaction is nil")
	}

	query := `
		INSERT INTO events (aggregate_id, transaction_id, aggregate_type, event_type, sequence_number, event_data, version)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	stmt, err := dbTx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	data, err := json.Marshal(event.EventData)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	_, err = stmt.ExecContext(ctx,
		event.AggregateID, event.TransactionID, event.AggregateType,
		event.EventType, event.SequenceNumber, string(data), event.Version)
	if err != nil {
		err = r.mapError(err)

		return fmt.Errorf("failed to exec statement: %w", err)
	}

	return nil
}

func (r *EventRepository) CreateBulkTx(ctx context.Context, dbTx *sql.Tx, events []model.Event) error {
	if dbTx == nil {
		return errors.New("transaction is nil")
	}

	// using pq.CopyIn to insert multiple rows at once leverage PostgreSQL COPY command
	query := pq.CopyIn("events", "aggregate_id", "transaction_id", "aggregate_type",
		"event_type", "sequence_number", "event_data", "version")

	stmt, err := dbTx.PrepareContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	for _, event := range events {
		data, err := json.Marshal(event.EventData)
		if err != nil {
			return fmt.Errorf("marshal error: %w", err)
		}

		_, err = stmt.ExecContext(ctx,
			event.AggregateID, event.TransactionID, event.AggregateType,
			event.EventType, event.SequenceNumber, string(data), event.Version)
		if err != nil {
			err = r.mapError(err)

			return fmt.Errorf("failed to exec statement: %w", err)
		}
	}

	// need to call ExecContext to execute the query
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		err = r.mapError(err)

		return fmt.Errorf("failed to exec flush statement: %w", err)
	}

	return nil
}

func (r *EventRepository) FindAllByTransactionID(ctx context.Context, transactionID string) ([]model.Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, sequence_number, event_data, version
		FROM events
		WHERE transaction_id = $1
	`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	rows, err := stmt.QueryContext(ctx, transactionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query statement: %w", err)
	}

	defer rows.Close()

	var events []model.Event

	for rows.Next() {
		var event model.Event

		err = rows.Scan(&event.ID, &event.AggregateID, &event.AggregateType,
			&event.EventType, &event.SequenceNumber, &event.EventData, &event.Version)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		events = append(events, event)
	}

	if len(events) == 0 {
		err := exception.ErrRecordNotFound
		err.MessageVars = map[string]interface{}{
			"name": "event",
		}

		return nil, fmt.Errorf("event not found: %w", err)
	}

	return events, nil
}

func (r *EventRepository) FindLastByAggregateID(ctx context.Context, aggregateID int64) (model.Event, error) {
	query := `
		SELECT id, aggregate_id, aggregate_type, event_type, sequence_number, event_data, version
		FROM events
		WHERE aggregate_id = $1
		ORDER BY sequence_number DESC
		LIMIT 1
	`

	stmt, err := r.db.PrepareContext(ctx, query)
	if err != nil {
		return model.Event{}, fmt.Errorf("failed to prepare statement: %w", err)
	}

	defer stmt.Close()

	row := stmt.QueryRowContext(ctx, aggregateID)

	var event model.Event
	err = row.Scan(&event.ID, &event.AggregateID, &event.AggregateType,
		&event.EventType, &event.SequenceNumber, &event.EventData, &event.Version)

	if err != nil && errors.Is(err, sql.ErrNoRows) {
		err := exception.ErrRecordNotFound
		err.MessageVars = map[string]interface{}{
			"name": "event",
		}

		return model.Event{}, fmt.Errorf("event not found: %w", err)
	}

	if err != nil {
		err = r.mapError(err)

		return model.Event{}, fmt.Errorf("failed to scan row: %w", err)
	}

	return event, nil
}
