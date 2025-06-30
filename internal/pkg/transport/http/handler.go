package http

import (
	"net/http"

	"github.com/go-kit/kit/endpoint"
	kithttp "github.com/go-kit/kit/transport/http"
)

var options = []kithttp.ServerOption{
	kithttp.ServerErrorEncoder(ErrorResponse),
}

// creates a generic http.Handler with the a given endpoint and a decode function.
// the response encoding function and the error encoding function are generic to all the JSON responses.
func MakeHTTPHandler(
	e endpoint.Endpoint,
	reqDecoder kithttp.DecodeRequestFunc,
	respEncoder kithttp.EncodeResponseFunc,
) http.Handler {
	return kithttp.NewServer(
		e,
		reqDecoder,
		respEncoder,
		options...,
	)
}

func MakeHandlerFunc(
	e endpoint.Endpoint,
	reqDecoder kithttp.DecodeRequestFunc,
	respEncoder kithttp.EncodeResponseFunc,
) http.HandlerFunc {
	return MakeHTTPHandler(e, reqDecoder, respEncoder).ServeHTTP
}
