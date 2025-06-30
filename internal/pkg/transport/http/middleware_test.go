//go:build unit

package http

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type structuredLog struct {
	Level      string          `json:"level"`
	Response   json.RawMessage `json:"response"`
	Request    json.RawMessage `json:"request"`
	Transport  string          `json:"transport"`
	Duration   float64         `json:"duration_ns"`
	Message    string          `json:"msg"`
	URL        string          `json:"url"`
	Method     string          `json:"method"`
	StatusCode int             `json:"status_code"`
	StackTrace string          `json:"stack_trace"`
}

func TestLoggingMiddleware(t *testing.T) {
	testCases := []struct {
		name               string
		statusCode         int
		expectedMsg        string
		expectedStatusCode int
		expectedLogLevel   string
	}{
		{
			name:               "success response",
			expectedMsg:        "successfully processing request",
			expectedStatusCode: http.StatusOK,
			expectedLogLevel:   "info",
		},
		{
			name:               "failure response",
			statusCode:         http.StatusUnprocessableEntity,
			expectedMsg:        "error processing request",
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedLogLevel:   "error",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				out      = new(bytes.Buffer)
				logger   = slog.New(slog.NewJSONHandler(out, nil))
				reqBody  = `{"foo":123}`
				respBody = `{"bar":"low"}`
				url      = "http://example.com/api/v1/users"
				method   = http.MethodPost
				req, _   = http.NewRequest(method, url, bytes.NewReader([]byte(reqBody)))
				handler  = func(respWriter http.ResponseWriter, req *http.Request) {
					if testCase.statusCode != 0 {
						respWriter.WriteHeader(testCase.statusCode)
					}
					respWriter.Write([]byte(respBody))
				}
				respRecorder = httptest.NewRecorder()
			)

			handlerWithLogging := LoggingMiddleware(logger)(http.HandlerFunc(handler))
			handlerWithLogging.ServeHTTP(respRecorder, req)

			var log structuredLog
			err := json.Unmarshal(out.Bytes(), &log)
			assert.Nil(t, err)

			assert.Equal(t, strings.ToUpper(testCase.expectedLogLevel), log.Level)
			assert.Equal(t, reqBody, unescapeUnquote(string(log.Request)))
			assert.Equal(t, respBody, unescapeUnquote(string(log.Response)))
			assert.Equal(t, "http", log.Transport)
			assert.NotZero(t, log.Duration)
			assert.Equal(t, testCase.expectedMsg, log.Message)
			assert.Equal(t, url, log.URL)
			assert.Equal(t, method, log.Method)
			assert.Equal(t, testCase.expectedStatusCode, log.StatusCode)
		})
	}
}

func TestLoggingMiddlewareMinifyJSON(t *testing.T) {
	var (
		out     = new(bytes.Buffer)
		logger  = slog.New(slog.NewJSONHandler(out, nil))
		reqBody = `{
				"foo":123
			}`
		respBody = `{
				"bar":"low"
			}`
		url     = "http://example.com/api/v1/users"
		method  = http.MethodPost
		req, _  = http.NewRequest(method, url, bytes.NewReader([]byte(reqBody)))
		handler = func(respWriter http.ResponseWriter, req *http.Request) {
			respWriter.WriteHeader(http.StatusOK)
			respWriter.Write([]byte(respBody))
		}
		respRecorder = httptest.NewRecorder()
	)

	handlerWithLogging := LoggingMiddleware(logger)(http.HandlerFunc(handler))
	handlerWithLogging.ServeHTTP(respRecorder, req)

	var log structuredLog
	err := json.Unmarshal(out.Bytes(), &log)
	assert.Nil(t, err)

	assert.Equal(t, `{"foo":123}`, unescapeUnquote(string(log.Request)))
	assert.Equal(t, `{"bar":"low"}`, unescapeUnquote(string(log.Response)))
}

func TestPanicRecovererMiddleware(t *testing.T) {
	var (
		handler = http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			// simulate panic
			panic("recover me!")
		})
		respRecorder = httptest.NewRecorder()
		req, _       = http.NewRequestWithContext(context.Background(), http.MethodGet, "http://example.com/api/v1/foo", nil)
		out          = new(bytes.Buffer)
		logger       = slog.New(slog.NewJSONHandler(out, nil))
	)

	assert.NotPanics(t, func() {
		Recoverer(logger)(handler).ServeHTTP(respRecorder, req)

		var log structuredLog
		err := json.Unmarshal(out.Bytes(), &log)
		assert.Nil(t, err)

		assert.NotEmpty(t, log.StackTrace)
	})
}

func unescapeUnquote(s string) string {
	s = strings.Trim(s, "\"")
	return strings.ReplaceAll(s, "\\", "")
}
