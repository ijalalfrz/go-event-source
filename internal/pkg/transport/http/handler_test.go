//go:build unit

package http

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler(t *testing.T) {
	dummyEndpoint := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	handler := MakeHTTPHandler(dummyEndpoint, DecodeRequest[dummyRequest], NoContentResponse)
	assert.Implements(t, (*http.Handler)(nil), handler)
}

func TestHandlerFunc(t *testing.T) {
	dummyEndpoint := func(ctx context.Context, req interface{}) (interface{}, error) {
		return nil, nil
	}

	handler := MakeHandlerFunc(dummyEndpoint, DecodeRequest[dummyRequest], NoContentResponse)
	assert.IsType(t, new(http.HandlerFunc), &handler)
}
