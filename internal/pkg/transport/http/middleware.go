package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/go-chi/cors"
	"github.com/ijalalfrz/go-event-source/internal/app/dto"
)

type MiddlewareFunc func(http.Handler) http.Handler

type loggingResponseWriter struct {
	http.ResponseWriter // original response writer
	body                []byte
	statusCode          int
}

func (r *loggingResponseWriter) Write(b []byte) (int, error) {
	size, err := r.ResponseWriter.Write(b) // write response using original http.ResponseWriter
	r.body = b                             // capture response body

	return size, err //nolint:wrapcheck
}

func (r *loggingResponseWriter) WriteHeader(statusCode int) {
	r.ResponseWriter.WriteHeader(statusCode) // write status code using original http.ResponseWriter
	r.statusCode = statusCode                // capture status code
}

func LoggingMiddleware(logger *slog.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			var (
				start             = time.Now()
				loggingRespWriter = loggingResponseWriter{ResponseWriter: respWriter}
				reqBytes          []byte
			)

			if req.Body != nil {
				defer req.Body.Close()
				reqBytes, _ = io.ReadAll(req.Body)
				req.Body = io.NopCloser(bytes.NewReader(reqBytes))
			}

			next.ServeHTTP(&loggingRespWriter, req)

			var (
				respBytes  = loggingRespWriter.body
				statusCode = loggingRespWriter.statusCode
			)

			if statusCode == 0 {
				// use default status code
				statusCode = http.StatusOK
			}

			attrs := []any{
				slog.String("type", "inbound"),
				slog.String("transport", "http"),
				slog.Duration("duration_ns", time.Duration(time.Since(start).Nanoseconds())),
				slog.String("url", req.URL.String()),
				slog.String("method", req.Method),
				slog.Int("status_code", statusCode),
				slog.String("request", string(compactJSON(reqBytes))),
				slog.String("response", string(compactJSON(respBytes))),
			}

			if statusCode >= http.StatusBadRequest {
				logger.ErrorContext(req.Context(), "error processing request", attrs...)
			} else {
				logger.InfoContext(req.Context(), "successfully processing request", attrs...)
			}
		})
	}
}

func compactJSON(jsonBytes []byte) []byte {
	if len(jsonBytes) == 0 {
		return jsonBytes
	}

	buffer := new(bytes.Buffer)

	err := json.Compact(buffer, jsonBytes)
	if err != nil {
		return jsonBytes
	}

	return buffer.Bytes()
}

func Recoverer(logger *slog.Logger) MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if err, _ := rvr.(error); errors.Is(err, http.ErrAbortHandler) {
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}

					logger.Error("panic occurred", slog.Any("message", rvr), slog.String("stack_trace", string(debug.Stack())))
					respWriter.WriteHeader(http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(respWriter, req)
		})
	}
}

// CORSMiddleware set CORS related headers.
func CORSMiddleware(allowedOrigins []string) func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins: allowedOrigins, // allow swagger
		AllowedMethods: []string{"GET", "POST", "PATCH", "PUT", "OPTIONS", "DELETE"},
		AllowedHeaders: []string{"Authorization", "Origin", "Content-Type", "X-Timestamp", "X-Transaction-Id"},
	})
}

func HeaderMiddleware() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(respWriter http.ResponseWriter, req *http.Request) {
			newReq, err := dto.RequestWithContext(req)
			if err != nil {
				ErrorResponse(req.Context(), err, respWriter)

				return
			}

			next.ServeHTTP(respWriter, newReq)
		})
	}
}
