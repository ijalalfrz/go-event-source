//go:build unit

package router

import (
	"net/http"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/ijalalfrz/go-event-source/internal/app/config"
	"github.com/ijalalfrz/go-event-source/internal/app/endpoint"
)

func TestConfigRoute(t *testing.T) {

	var origins []string
	origins = append(origins, "http://localhost:3000")

	cfg := config.Config{
		HTTP: config.HTTP{
			AllowedOrigin: origins,
		},
	}

	router := MakeHTTPRouter(
		endpoint.Endpoint{
			Account:     endpoint.Account{},
			Transaction: endpoint.Transaction{},
		},
		cfg,
	)

	testCases := []struct {
		name        string
		method      string
		path        string
		shouldMatch bool
	}{
		{
			name:        "Healthcheck",
			method:      http.MethodGet,
			path:        "/health",
			shouldMatch: true,
		},
		{
			name:        "Create Account",
			method:      http.MethodPost,
			path:        "/accounts",
			shouldMatch: true,
		},
		{
			name:        "Get Account",
			method:      http.MethodGet,
			path:        "/accounts/1",
			shouldMatch: true,
		},
		{
			name:        "Create Transfer",
			method:      http.MethodPost,
			path:        "/transactions",
			shouldMatch: true,
		},
	}

	chiCtx := chi.NewRouteContext()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			match := router.Match(chiCtx, testCase.method, testCase.path)

			if testCase.shouldMatch && !match {
				t.Errorf("Route doesn't match!")
			} else if !testCase.shouldMatch && match {
				t.Error("Unexpected route match!")
			}
		})
	}
}
