package http

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-chi/render"
)

func init() { //nolint:gochecknoinits
	render.Decode = Decoder
}

func Decoder(req *http.Request, val interface{}) error {
	if req.Body == nil {
		return nil
	}

	var err error

	if render.GetRequestContentType(req) == render.ContentTypeJSON {
		err = render.DecodeJSON(req.Body, val)
		if err != nil && errors.Is(err, io.EOF) {
			err = nil // not all requests have body
		}
	}

	return err
}

func DecodeRequest[T any, PT interface {
	render.Binder
	*T
}](_ context.Context, req *http.Request) (interface{}, error) {
	binder := PT(new(T))

	err := render.Bind(req, binder)
	if err != nil {
		return nil, fmt.Errorf("http bind request: %w", err)
	}

	return binder, nil
}
