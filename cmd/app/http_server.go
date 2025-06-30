package app

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/ijalalfrz/go-event-source/internal/app/config"
	"github.com/ijalalfrz/go-event-source/internal/app/endpoint"
	"github.com/ijalalfrz/go-event-source/internal/app/repository"
	"github.com/ijalalfrz/go-event-source/internal/app/router"
	"github.com/ijalalfrz/go-event-source/internal/app/service"
	"github.com/ijalalfrz/go-event-source/internal/pkg/db"
	"github.com/ijalalfrz/go-event-source/internal/pkg/lang"
	"github.com/ijalalfrz/go-event-source/internal/pkg/logger"
	"github.com/spf13/cobra"
)

var timeout = 30 * time.Second

var httpServerCmd = &cobra.Command{
	Use:   "http",
	Short: "Serve incoming requests from REST HTTP/JSON API",
	Run: func(_ *cobra.Command, _ []string) {
		slog.Debug("command line flags", slog.String("config_path", cfgFilePath))
		cfg := config.MustInitConfig(cfgFilePath)

		logger.InitStructuredLogger(cfg.LogLevel)

		runHTTPServer(cfg)
	},
}

func runHTTPServer(cfg config.Config) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var waitGroup sync.WaitGroup

	slog.InfoContext(ctx, "starting...", slog.String("log_level", string(cfg.LogLevel)))

	// Starts the server in a go routine
	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		startHTTPServer(ctx, cfg)
	}()

	sigChannel := make(chan os.Signal, 1)
	signal.Notify(sigChannel, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)

	select {
	case sig := <-sigChannel:
		cancel()
		slog.InfoContext(ctx, "received OS signal. Exiting...", slog.String("signal", sig.String()))
	case <-ctx.Done():
		slog.ErrorContext(ctx, "failed to start HTTP server")
	}

	waitGroup.Wait()
	slog.Info("All servers stopped")
}

// startHTTPServer loads config and starts HTTP server.
func startHTTPServer(ctx context.Context, cfg config.Config) {
	var waitGroup sync.WaitGroup

	waitGroup.Add(1)

	go func() {
		defer waitGroup.Done()
		startMainApp(ctx, cfg)
	}()

	if cfg.HTTP.PprofEnabled {
		waitGroup.Add(1)

		go func() {
			defer waitGroup.Done()
			startPprof(ctx, cfg)
		}()
	}

	waitGroup.Wait()
}

func startMainApp(ctx context.Context, cfg config.Config) {
	lang.SetSupportedLanguages(cfg.Locales.SupportedLanguages)
	lang.SetBasePath(cfg.Locales.BasePath)

	endpts := makeEndpoints(cfg)

	router := router.MakeHTTPRouter(
		endpts,
		cfg,
	)

	server := &http.Server{
		Handler:      router,
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		WriteTimeout: cfg.HTTP.Timeout,
		ReadTimeout:  cfg.HTTP.Timeout,
	}

	slog.Info("running HTTP server...", slog.Int("port", cfg.HTTP.Port))

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()

	// shutdown ctx
	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown HTTP server", slog.String("error", err.Error()))
	}

	slog.Info("HTTP server gracefully stopped")
}

func makeEndpoints(cfg config.Config) endpoint.Endpoint {
	dbConn := db.InitDB(cfg)

	// init all repo
	accountRepository := repository.NewAccountRepository(dbConn)
	eventRepository := repository.NewEventRepository(dbConn)

	return endpoint.Endpoint{
		Account:     makeAccountEndpoints(accountRepository, eventRepository, cfg),
		Transaction: makeTransactionEndpoints(accountRepository, eventRepository, cfg),
	}
}

func makeAccountEndpoints(accountRepository *repository.AccountRepository,
	eventRepository *repository.EventRepository, cfg config.Config,
) endpoint.Account {
	accountSvc := service.NewAccountService(accountRepository, eventRepository,
		cfg.RequestTimeThreshold, cfg.EventVersion)

	return endpoint.NewAccountEndpoint(accountSvc)
}

func makeTransactionEndpoints(accountRepository *repository.AccountRepository,
	eventRepository *repository.EventRepository, cfg config.Config,
) endpoint.Transaction {
	transactionSvc := service.NewTransactionService(accountRepository, eventRepository,
		cfg.RequestTimeThreshold, cfg.EventVersion)

	return endpoint.NewTransactionEndpoint(transactionSvc)
}

func startPprof(ctx context.Context, cfg config.Config) {
	// manually register pprof handlers with custom path.
	http.HandleFunc("/internal/pprof/", pprof.Index)
	http.HandleFunc("/internal/pprof/cmdline", pprof.Cmdline)
	http.HandleFunc("/internal/pprof/profile", pprof.Profile)
	http.HandleFunc("/internal/pprof/symbol", pprof.Symbol)
	http.HandleFunc("/internal/pprof/trace", pprof.Trace)
	slog.Info("running pprof...", slog.Int("port", cfg.HTTP.PprofPort))

	server := &http.Server{
		Addr:              fmt.Sprintf("localhost:%d", cfg.HTTP.PprofPort),
		ReadHeaderTimeout: cfg.HTTP.Timeout,
	}

	// limit access by binding port to localhost only.
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("pprof server error", slog.String("error", err.Error()))
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		slog.Error("failed to shutdown pprof server", slog.String("error", err.Error()))
	}

	slog.Info("pprof server stopped")
}
