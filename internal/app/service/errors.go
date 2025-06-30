package service

import (
	"net/http"

	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/ijalalfrz/go-event-source/internal/pkg/lang"
)

// list of error sentinel in service.
var ErrSourceAccountNotFound = exception.ApplicationError{
	Localizable: lang.Localizable{
		MessageID: "errors.source_account_not_found",
		Message:   "source account not found",
	},
	StatusCode: http.StatusNotFound,
}

var ErrDestinationAccountNotFound = exception.ApplicationError{
	Localizable: lang.Localizable{
		MessageID: "errors.destination_account_not_found",
		Message:   "destination account not found",
	},
	StatusCode: http.StatusNotFound,
}

var ErrInsufficientBalance = exception.ApplicationError{
	Localizable: lang.Localizable{
		MessageID: "errors.insufficient_balance",
		Message:   "insufficient balance",
	},
	StatusCode: http.StatusBadRequest,
}

var ErrInvalidRequestTime = exception.ApplicationError{
	Localizable: lang.Localizable{
		MessageID: "errors.invalid_request_time",
		Message:   "invalid request time",
	},
	StatusCode: http.StatusBadRequest,
}

var ErrIdempotency = exception.ApplicationError{
	Localizable: lang.Localizable{
		MessageID: "errors.idempotency",
		Message:   "transaction id already used by another operation",
	},
	StatusCode: http.StatusConflict,
}

var ErrSourceAndDestinationAccountSame = exception.ApplicationError{
	Localizable: lang.Localizable{
		MessageID: "errors.source_and_destination_account_same",
		Message:   "source and destination account cannot be the same",
	},
	StatusCode: http.StatusConflict,
}

var ErrAccountAlreadyExists = exception.ApplicationError{
	Localizable: lang.Localizable{
		MessageID: "errors.account_already_exists",
		Message:   "account already exists",
	},
	StatusCode: http.StatusConflict,
}
