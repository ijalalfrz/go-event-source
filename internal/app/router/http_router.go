package router

import (
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/ijalalfrz/go-event-source/internal/app/config"
	"github.com/ijalalfrz/go-event-source/internal/app/dto"
	"github.com/ijalalfrz/go-event-source/internal/app/endpoint"
	httptransport "github.com/ijalalfrz/go-event-source/internal/pkg/transport/http"
)

// MakeHTTPRouter builds the HTTP router with all the service endpoints.
func MakeHTTPRouter(
	endpts endpoint.Endpoint,
	cfg config.Config,
) *chi.Mux {
	// Initialize Router
	router := chi.NewRouter()

	router.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	router.Route("/", func(router chi.Router) {
		router.Use(
			httptransport.LoggingMiddleware(slog.Default()),
			httptransport.CORSMiddleware(cfg.HTTP.AllowedOrigin),
			httptransport.Recoverer(slog.Default()),
			render.SetContentType(render.ContentTypeJSON),
		)

		router.Route("/accounts", func(router chi.Router) {
			routerWithHeader := router.With(httptransport.HeaderMiddleware())
			routerWithHeader.Post("/", httptransport.MakeHandlerFunc(
				endpts.Account.Create,
				httptransport.DecodeRequest[dto.CreateAccountRequest],
				httptransport.CreatedResponse,
			))
			router.Get("/{id}", httptransport.MakeHandlerFunc(
				endpts.Account.Get,
				httptransport.DecodeRequest[dto.GetAccountRequest],
				httptransport.ResponseWithBody,
			))
		})

		router.Route("/transactions", func(router chi.Router) {
			routerWithHeader := router.With(httptransport.HeaderMiddleware())
			routerWithHeader.Post("/", httptransport.MakeHandlerFunc(
				endpts.Transaction.Transfer,
				httptransport.DecodeRequest[dto.CreateTransferRequest],
				httptransport.NoContentResponse,
			))
		})
	})

	return router
}
