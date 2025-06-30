package endpoint

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/ijalalfrz/go-event-source/internal/app/dto"
)

type TransactionService interface {
	Transfer(ctx context.Context, req dto.CreateTransferRequest) error
}

func NewTransactionEndpoint(service TransactionService) Transaction {
	return Transaction{
		Transfer: makeTransferEndpoint(service),
	}
}

func makeTransferEndpoint(service TransactionService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*dto.CreateTransferRequest)
		if !ok {
			return nil, fmt.Errorf("transaction transfer request type: %w", ErrInvalidType)
		}

		if err := service.Transfer(ctx, *req); err != nil {
			return nil, fmt.Errorf("transaction service: %w", err)
		}

		return nil, nil
	}
}
