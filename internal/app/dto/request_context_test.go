//go:build unit

package dto

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRequestContext(t *testing.T) {
	var (
		language      = "en"
		timestamp     = "2025-01-01T00:00:00Z"
		transactionID = "1234567890"
	)

	req, err := http.NewRequestWithContext(context.Background(), "GET", "/foo", nil)
	assert.NoError(t, err)

	req.Header.Add("Accept-Language", language)
	req.Header.Add("X-TIMESTAMP", timestamp)
	req.Header.Add("X-TRANSACTION-ID", transactionID)

	out, err := RequestWithContext(req)
	assert.NoError(t, err)

	reqContext, ok := RequestFromContext(out.Context())
	assert.True(t, ok)

	assert.Equal(t, language, reqContext.Language)
	assert.Equal(t, timestamp, reqContext.Timestamp.Format(time.RFC3339))
	assert.Equal(t, transactionID, reqContext.TransactionID)
}
