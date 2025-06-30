//go:build unit

package exception

import (
	"errors"
	"net/http"
	"testing"

	"github.com/ijalalfrz/go-event-source/internal/pkg/lang"
	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	err := ApplicationError{
		Localizable: lang.Localizable{Message: "error occured"},
		StatusCode:  500,
	}

	t.Run("Implements error interface", func(t *testing.T) {
		assert.Implements(t, (*error)(nil), err)
		assert.Equal(t, "error occured", err.Error())
	})

	t.Run("Error without cause", func(t *testing.T) {
		err.Cause = nil

		assert.Implements(t, (*error)(nil), err)
		assert.Equal(t, "error occured", err.Error())
	})

	t.Run("ErrorCode() returns StatusCode", func(t *testing.T) {
		assert.Equal(t, err.StatusCode, err.ErrorCode())
	})
}

func TestGetHTTPStatusCodeByErr(t *testing.T) {
	testCases := []struct {
		description        string
		err                error
		expectedStatusCode int
	}{
		{
			description:        "Validation error",
			err:                ApplicationError{StatusCode: http.StatusBadRequest},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			description:        "Internal server error",
			err:                errors.New("connection lost"),
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.description, func(t *testing.T) {
			assert.Equal(t, testCase.expectedStatusCode, GetHTTPStatusCodeByErr(testCase.err))
		})
	}
}
