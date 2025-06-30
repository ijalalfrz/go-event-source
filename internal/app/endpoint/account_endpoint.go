package endpoint

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/endpoint"
	"github.com/ijalalfrz/go-event-source/internal/app/dto"
)

type AccountService interface {
	CreateAccount(ctx context.Context, req dto.CreateAccountRequest) error
	GetAccount(ctx context.Context, req dto.GetAccountRequest) (dto.AccountResponse, error)
}

func NewAccountEndpoint(service AccountService) Account {
	return Account{
		Create: makeCreateAccountEndpoint(service),
		Get:    makeGetAccountEndpoint(service),
	}
}

// makeCreateAccountEndpoint is a helper function to create a new account endpoint POST /accounts.
func makeCreateAccountEndpoint(service AccountService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*dto.CreateAccountRequest)
		if !ok {
			return nil, fmt.Errorf("account create request type: %w", ErrInvalidType)
		}

		if err := service.CreateAccount(ctx, *req); err != nil {
			return nil, fmt.Errorf("account service: %w", err)
		}

		return nil, nil
	}
}

// makeGetAccountEndpoint is a helper function to get an account endpoint GET /accounts/{id}.
func makeGetAccountEndpoint(service AccountService) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req, ok := request.(*dto.GetAccountRequest)
		if !ok {
			return nil, fmt.Errorf("account get request type: %w", ErrInvalidType)
		}

		account, err := service.GetAccount(ctx, *req)
		if err != nil {
			return nil, fmt.Errorf("account service: %w", err)
		}

		return account, nil
	}
}
