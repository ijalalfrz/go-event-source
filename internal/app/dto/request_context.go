package dto

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/ijalalfrz/go-event-source/internal/pkg/lang"
)

type RequestContext struct {
	Signature     string    `mapstructure:"signature"`
	Language      string    `mapstructure:"language"`
	Timestamp     time.Time `mapstructure:"timestamp"`
	TransactionID string    `mapstructure:"transaction_id"`
}

type contextKey string

// requestContextKey is the context.Context key to store the request context.
var requestContextKey = contextKey("request_context")

func RequestWithContext(req *http.Request) (*http.Request, error) {
	var reqContext RequestContext

	reqContext.Language = getLanguage(req)
	reqContext.Signature = getSignature(req)
	reqContext.TransactionID = getTransactionID(req)

	timestamp, err := getTimestamp(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get X-TIMESTAMP in header: %w", err)
	}

	reqContext.Timestamp = timestamp

	ctx := context.WithValue(req.Context(), requestContextKey, reqContext)

	return req.WithContext(ctx), nil
}

func RequestFromContext(ctx context.Context) (RequestContext, bool) {
	reqContext, ok := ctx.Value(requestContextKey).(RequestContext)

	return reqContext, ok
}

func getLanguage(req *http.Request) string {
	return req.Header.Get("Accept-Language")
}

func getSignature(req *http.Request) string {
	return req.Header.Get("X-Signature")
}

func getTimestamp(req *http.Request) (time.Time, error) {
	timestampStr := req.Header.Get("X-Timestamp")
	if timestampStr == "" {
		err := exception.ApplicationError{
			Localizable: lang.Localizable{
				Message: "X-TIMESTAMP is required in header",
			},
			StatusCode: http.StatusBadRequest,
		}

		return time.Time{}, err
	}

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse timestamp: %w", err)
	}

	return timestamp, nil
}

func getTransactionID(req *http.Request) string {
	return req.Header.Get("X-Transaction-Id")
}
