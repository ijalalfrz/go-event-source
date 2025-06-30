//go:build unit

package service

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/ijalalfrz/go-event-source/internal/app/dto"
	"github.com/stretchr/testify/assert"
)

// Helper function to create context with request context using the actual dto package
func createContextWithRequestContext(reqContext dto.RequestContext) context.Context {
	// Create a context with the request context
	ctx := context.Background()

	req, _ := createMockRequest(reqContext)
	if req != nil {
		return req.Context()
	}

	key := contextKey("request_context")
	return context.WithValue(ctx, key, reqContext)
}

// contextKey matches the one in dto package
type contextKey string

// Helper function to create a mock HTTP request with the request context
func createMockRequest(reqContext dto.RequestContext) (*http.Request, error) {
	req, err := http.NewRequest("GET", "/test", nil)
	if err != nil {
		return nil, err
	}

	// Set headers that RequestWithContext would read
	if reqContext.Signature != "" {
		req.Header.Set("X-SIGNATURE", reqContext.Signature)
	}
	if reqContext.Language != "" {
		req.Header.Set("Accept-Language", reqContext.Language)
	}
	if reqContext.TransactionID != "" {
		req.Header.Set("X-TRANSACTION-ID", reqContext.TransactionID)
	}
	if !reqContext.Timestamp.IsZero() {
		req.Header.Set("X-TIMESTAMP", reqContext.Timestamp.Format(time.RFC3339))
	}

	return dto.RequestWithContext(req)
}

func TestGetRequestContext(t *testing.T) {
	t.Run("success_with_valid_context", func(t *testing.T) {
		reqContext := dto.RequestContext{
			Signature:     "test-signature",
			Language:      "en",
			Timestamp:     time.Now(),
			TransactionID: "tx-12345",
		}
		ctx := createContextWithRequestContext(reqContext)

		result, err := getRequestContext(ctx, 30*time.Second)

		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "test-signature", result.Signature)
		assert.Equal(t, "en", result.Language)
		assert.Equal(t, "tx-12345", result.TransactionID)

		timeDiff := time.Since(result.Timestamp)
		assert.True(t, timeDiff >= -1*time.Second && timeDiff <= 1*time.Second,
			"timestamp should be within 1 second of now")
	})

	t.Run("error_missing_request_context", func(t *testing.T) {
		ctx := context.Background()

		result, err := getRequestContext(ctx, 30*time.Second)

		assert.Error(t, err)
		assert.NotNil(t, err)
		assert.Empty(t, result)
	})

	t.Run("error_expired_request_positive_threshold", func(t *testing.T) {
		reqContext := dto.RequestContext{
			Signature:     "test-signature",
			Language:      "en",
			Timestamp:     time.Now().Add(-35 * time.Second), // 35 seconds ago
			TransactionID: "tx-12345",
		}
		ctx := createContextWithRequestContext(reqContext)

		result, err := getRequestContext(ctx, 30*time.Second)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRequestTime)
		assert.Empty(t, result)
	})

	t.Run("error_expired_request_negative_threshold", func(t *testing.T) {
		reqContext := dto.RequestContext{
			Signature:     "test-signature",
			Language:      "en",
			Timestamp:     time.Now().Add(35 * time.Second), // 35 seconds in future
			TransactionID: "tx-12345",
		}
		ctx := createContextWithRequestContext(reqContext)

		result, err := getRequestContext(ctx, 30*time.Second)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRequestTime)
		assert.Empty(t, result)
	})

	t.Run("success_at_threshold_boundary", func(t *testing.T) {
		reqContext := dto.RequestContext{
			Signature:     "test-signature",
			Language:      "en",
			Timestamp:     time.Now().Add(-29 * time.Second), // Just under threshold
			TransactionID: "tx-12345",
		}
		ctx := createContextWithRequestContext(reqContext)

		result, err := getRequestContext(ctx, 30*time.Second)

		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.Equal(t, "test-signature", result.Signature)
		assert.Equal(t, "en", result.Language)
		assert.Equal(t, "tx-12345", result.TransactionID)
	})

}

func TestGetRequestContext_EdgeCases(t *testing.T) {
	t.Run("nil_context", func(t *testing.T) {
		// This should panic or return error
		assert.Panics(t, func() {
			getRequestContext(nil, 30*time.Second)
		})
	})

	t.Run("context_with_wrong_type", func(t *testing.T) {
		// Context with wrong type value
		ctx := context.WithValue(context.Background(), contextKey("request_context"), "wrong-type")

		result, err := getRequestContext(ctx, 30*time.Second)

		assert.Error(t, err)
		assert.Empty(t, result)
	})

	t.Run("context_with_nil_value", func(t *testing.T) {
		// Context with nil value
		ctx := context.WithValue(context.Background(), contextKey("request_context"), nil)

		result, err := getRequestContext(ctx, 30*time.Second)

		assert.Error(t, err)
		assert.Empty(t, result)
	})
}

func TestGetRequestContext_TimePrecision(t *testing.T) {

	t.Run("very_old_request", func(t *testing.T) {
		// Test with very old request
		threshold := 1 * time.Second

		reqContext := dto.RequestContext{
			Signature:     "test-signature",
			Language:      "en",
			Timestamp:     time.Now().Add(-1 * time.Hour), // 1 hour ago
			TransactionID: "tx-12345",
		}
		ctx := createContextWithRequestContext(reqContext)

		result, err := getRequestContext(ctx, threshold)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRequestTime)
		assert.Empty(t, result)
	})

	t.Run("future_request", func(t *testing.T) {
		// Test with future request
		threshold := 1 * time.Second

		reqContext := dto.RequestContext{
			Signature:     "test-signature",
			Language:      "en",
			Timestamp:     time.Now().Add(1 * time.Hour), // 1 hour in future
			TransactionID: "tx-12345",
		}
		ctx := createContextWithRequestContext(reqContext)

		result, err := getRequestContext(ctx, threshold)

		assert.Error(t, err)
		assert.ErrorIs(t, err, ErrInvalidRequestTime)
		assert.Empty(t, result)
	})
}
