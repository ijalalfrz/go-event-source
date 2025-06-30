package exception

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/ijalalfrz/go-event-source/internal/pkg/lang"
)

const (
	CodeBadRequest    = http.StatusBadRequest
	CodeNotFound      = http.StatusNotFound
	CodeUnprocessable = http.StatusUnprocessableEntity
	CodeInternal      = http.StatusInternalServerError
	CodeUnauthorized  = http.StatusUnauthorized
	CodeForbidden     = http.StatusForbidden
	CodeConflict      = http.StatusConflict
)

var (
	ErrRecordNotFound = ApplicationError{
		Localizable: lang.Localizable{
			MessageVars: map[string]interface{}{"name": "Resource"},
			MessageID:   "errors.record_not_found",
			Message:     "record not found",
		},
		StatusCode: CodeNotFound,
	}

	ErrRecordNotUnique = ApplicationError{
		Localizable: lang.Localizable{
			MessageID: "errors.record_not_unique",
			Message:   "record not unique",
		},
		StatusCode: CodeUnprocessable,
	}
	ErrUnauthorized = ApplicationError{
		Localizable: lang.Localizable{
			MessageID: "errors.request_unauthorized",
			Message:   "request unauthorized",
		},
		StatusCode: CodeUnauthorized,
	}

	ErrConflict = ApplicationError{
		Localizable: lang.Localizable{
			MessageID: "errors.record_already_exist",
			Message:   "record already exist",
		},
		StatusCode: CodeConflict,
	}
)

// ApplicationError handles application level errors.
type ApplicationError struct {
	lang.Localizable
	StatusCode int
	Cause      error
}

// Error interface implementation.
func (e ApplicationError) Error() string {
	if e.Cause == nil {
		return e.Localizable.Message
	}

	return fmt.Sprintf("%s: %s", e.Localizable.Message, e.Cause)
}

// ErrorCode returns error code for an application error.
func (e ApplicationError) ErrorCode() int {
	return e.StatusCode
}

func (e ApplicationError) Is(err error) bool {
	var appErr ApplicationError

	ok := errors.As(err, &appErr)
	if !ok {
		return false
	}

	return appErr.StatusCode == e.StatusCode &&
		appErr.Localizable.Message == e.Localizable.Message &&
		e.Cause == appErr.Cause
}

// GetHTTPStatusCodeByErr returns HTTP status code mapping from error.
func GetHTTPStatusCodeByErr(err error) int {
	var appErr ApplicationError
	if errors.As(err, &appErr) {
		return appErr.ErrorCode()
	}

	// default to internal server error
	return http.StatusInternalServerError
}
