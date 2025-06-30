package endpoint

import (
	"github.com/go-kit/kit/endpoint"
	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/ijalalfrz/go-event-source/internal/pkg/lang"
)

// ErrInvalidType invalid type of request.
var ErrInvalidType = exception.ApplicationError{
	Localizable: lang.Localizable{
		Message: "invalid type",
	},
	StatusCode: exception.CodeBadRequest,
}

type Account struct {
	Create endpoint.Endpoint
	Get    endpoint.Endpoint
}

type Transaction struct {
	Transfer endpoint.Endpoint
}

type Endpoint struct {
	Account
	Transaction
}
