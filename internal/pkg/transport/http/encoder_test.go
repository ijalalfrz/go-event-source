//go:build unit

package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ijalalfrz/go-event-source/internal/pkg/exception"
	"github.com/ijalalfrz/go-event-source/internal/pkg/lang"
	"github.com/stretchr/testify/assert"
)

func TestEncodeError(t *testing.T) {
	resp := httptest.NewRecorder()
	ErrorResponse(context.Background(), errors.New("internal error"), resp)

	assert.Equal(t, http.StatusInternalServerError, resp.Code)
	assert.JSONEq(t, `{"error": "internal error"}`, resp.Body.String())
}

func TestEncodeErrorCustomError(t *testing.T) {
	lang.SetBasePath("../../../../resources/locales")
	resp := httptest.NewRecorder()
	err := exception.ApplicationError{
		StatusCode:  http.StatusBadRequest,
		Localizable: lang.Localizable{Message: "invalid request"},
	}
	ErrorResponse(context.Background(), err, resp)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	assert.JSONEq(t, `{"error": "invalid request"}`, resp.Body.String())
}

func TestEncodeJSONResponse(t *testing.T) {
	resp := httptest.NewRecorder()
	err := ResponseWithBody(context.Background(), resp, map[string]string{"foo": "bar"})

	assert.Nil(t, err)
	assert.Equal(t, http.StatusOK, resp.Code)
	assert.Equal(t, "application/json; charset=utf-8", resp.Result().Header.Get("Content-Type"))
	assert.JSONEq(t, `{"foo": "bar"}`, resp.Body.String())
}

func TestNoContentResponse(t *testing.T) {
	resp := httptest.NewRecorder()
	err := NoContentResponse(context.Background(), resp, nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusNoContent, resp.Code)
}

func TestCreatedResponse(t *testing.T) {
	resp := httptest.NewRecorder()
	err := CreatedResponse(context.Background(), resp, nil)

	assert.Nil(t, err)
	assert.Equal(t, http.StatusCreated, resp.Code)
}
