//go:build unit

package http

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type dummyRequest struct {
	Foo string `json:"foo"`
	Bar string // is a query param
}

func (d *dummyRequest) Bind(r *http.Request) error {
	d.Bar = r.URL.Query().Get("bar")

	return nil
}

func TestBindValidRequest(t *testing.T) {
	ctx := context.Background()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "/dummy/url?bar=123", strings.NewReader(`{"foo": "heho"}`))
	assert.Nil(t, err)

	request.Header.Add("Content-Type", "application/json")

	decoded, err := DecodeRequest[dummyRequest](ctx, request)
	assert.Nil(t, err)

	foo, ok := decoded.(*dummyRequest)
	assert.True(t, ok)

	assert.Equal(t, "heho", foo.Foo)
	assert.Equal(t, "123", foo.Bar)
}

func TestBindInvalidRequest(t *testing.T) {
	ctx := context.Background()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "/dummy/url?bar=123", strings.NewReader(`not a json`))
	assert.Nil(t, err)

	request.Header.Add("Content-Type", "application/json")

	_, err = DecodeRequest[dummyRequest](ctx, request)
	assert.Contains(t, err.Error(), "http bind request")
}

func TestBindEmptyRequestBody(t *testing.T) {
	ctx := context.Background()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "/dummy/url?bar=123", http.NoBody)
	assert.Nil(t, err)

	request.Header.Add("Content-Type", "application/json")

	decoded, err := DecodeRequest[dummyRequest](ctx, request)
	assert.Nil(t, err)

	foo, ok := decoded.(*dummyRequest)
	assert.True(t, ok)

	assert.Equal(t, "", foo.Foo)
	assert.Equal(t, "123", foo.Bar)
}

func TestBindBodyNilRequest(t *testing.T) {
	ctx := context.Background()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, "/dummy/url?bar=123", nil)
	assert.Nil(t, err)

	_, err = DecodeRequest[dummyRequest](ctx, request)
	assert.Nil(t, err)
	assert.Equal(t, nil, nil)
}
