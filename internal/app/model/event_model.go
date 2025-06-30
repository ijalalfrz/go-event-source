package model

import (
	"time"
)

type AggregateType string

const (
	AggregateTypeAccount AggregateType = "account"
)

type EventType string

const (
	EventTypeInitBalance     EventType = "init_balance"
	EventTypeDepositReceived EventType = "deposit_received"
	EventTypeDebitBalance    EventType = "balance_debited"
	EventTypeCreditBalance   EventType = "balance_credited"
)

type Event struct {
	ID             int64         `json:"id"`
	TransactionID  string        `json:"transaction_id"`
	SequenceNumber int64         `json:"sequence_number"`
	AggregateID    int64         `json:"aggregate_id"`
	AggregateType  AggregateType `json:"aggregate_type"`
	EventType      EventType     `json:"event_type"`
	EventData      interface{}   `json:"event_data"`
	Version        string        `json:"version"`
	CreatedAt      time.Time     `json:"created_at"`
}
