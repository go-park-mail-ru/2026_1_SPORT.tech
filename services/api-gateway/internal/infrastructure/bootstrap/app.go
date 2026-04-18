package bootstrap

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/health"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
	loggerinfra "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/logger"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/metrics"
)

type App struct {
	cfg        config.Config
	logger     *slog.Logger
	httpServer *http.Server
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	logger := loggerinfra.New(cfg.ServiceName)
	metricsSet := metrics.New(cfg.ServiceName)

	gatewayHandler, err := httpgateway.NewMux(ctx, httpgateway.DownstreamEndpoints{
		AuthGRPCEndpoint:    cfg.Downstream.AuthGRPCEndpoint,
		ProfileGRPCEndpoint: cfg.Downstream.ProfileGRPCEndpoint,
		ContentGRPCEndpoint: cfg.Downstream.ContentGRPCEndpoint,
	})
	if err != nil {
		return nil, err
	}

	healthChecker := health.NewGRPCChecker([]health.Dependency{
		{Name: "auth-service", Endpoint: cfg.Downstream.AuthGRPCEndpoint},
		{Name: "profile-service", Endpoint: cfg.Downstream.ProfileGRPCEndpoint},
		{Name: "content-service", Endpoint: cfg.Downstream.ContentGRPCEndpoint},
	})

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", metricsSet.Handler())
	httpMux.Handle("/healthz", health.NewHandler(cfg.ServiceName, healthChecker))
	httpMux.Handle("/api/openapi/auth.swagger.json", httpgateway.OpenAPIHandler(cfg.OpenAPI.AuthFilePath))
	httpMux.Handle("/api/openapi/profile.swagger.json", httpgateway.OpenAPIHandler(cfg.OpenAPI.ProfileFilePath))
	httpMux.Handle("/api/openapi/content.swagger.json", httpgateway.OpenAPIHandler(cfg.OpenAPI.ContentFilePath))
	httpMux.Handle("/api/", http.StripPrefix("/api", gatewayHandler))

	httpServer := &http.Server{
		Addr:              cfg.Server.HTTPAddress(),
		Handler:           metricsSet.HTTPMiddleware(httpMux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      20 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		cfg:        cfg,
		logger:     logger,
		httpServer: httpServer,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		app.logger.Info("starting HTTP gateway", "addr", app.cfg.Server.HTTPAddress())
		if err := app.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		app.logger.Info("shutdown requested")
		return app.Shutdown()
	case err := <-errCh:
		_ = app.Shutdown()
		return err
	}
}

func (app *App) Shutdown() error {
	shutdownTimeout, err := app.cfg.Server.ShutdownTimeoutDuration()
	if err != nil {
		shutdownTimeout = 10 * time.Second
	}

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := app.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}

	return nil
}
