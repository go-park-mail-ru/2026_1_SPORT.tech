package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/config"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/health"
	grpcserverinfra "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/grpcserver"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
	loggerinfra "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/logger"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/metrics"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type App struct {
	cfg        config.Config
	logger     *slog.Logger
	grpcServer *grpcserverinfra.Server
	httpServer *http.Server
	conns      []*grpc.ClientConn
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	logger := loggerinfra.New(cfg.ServiceName)
	metricsSet := metrics.New(cfg.ServiceName)

	authConn, err := grpc.DialContext(ctx, cfg.Downstream.AuthGRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("dial auth-service: %w", err)
	}
	profileConn, err := grpc.DialContext(ctx, cfg.Downstream.ProfileGRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = authConn.Close()
		return nil, fmt.Errorf("dial profile-service: %w", err)
	}
	contentConn, err := grpc.DialContext(ctx, cfg.Downstream.ContentGRPCEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = authConn.Close()
		_ = profileConn.Close()
		return nil, fmt.Errorf("dial content-service: %w", err)
	}

	gatewayService := grpcadapter.NewServer(
		authv1.NewAuthServiceClient(authConn),
		profilev1.NewProfileServiceClient(profileConn),
		contentv1.NewContentServiceClient(contentConn),
	)
	grpcServer, err := grpcserverinfra.New(cfg.Server.GRPCListenAddress(), gatewayService, gatewayService, gatewayService)
	if err != nil {
		_ = authConn.Close()
		_ = profileConn.Close()
		_ = contentConn.Close()
		return nil, err
	}

	gatewayHandler, err := httpgateway.NewMux(ctx, gatewayService, gatewayService, gatewayService)
	if err != nil {
		_ = authConn.Close()
		_ = profileConn.Close()
		_ = contentConn.Close()
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
	httpMux.Handle("/api/docs", http.RedirectHandler("/api/docs/", http.StatusMovedPermanently))
	httpMux.Handle("/api/docs/", httpgateway.DocsHandler("/api/openapi/gateway.swagger.json"))
	httpMux.Handle("/api/openapi/gateway.swagger.json", httpgateway.GatewayOpenAPIHandler(cfg.OpenAPI.GatewayFilePath))
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
		grpcServer: grpcServer,
		httpServer: httpServer,
		conns:      []*grpc.ClientConn{authConn, profileConn, contentConn},
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		app.logger.Info("starting gRPC gateway", "addr", app.cfg.Server.GRPCListenAddress())
		if err := app.grpcServer.Serve(); err != nil {
			errCh <- err
		}
	}()

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

	if err := app.grpcServer.Shutdown(ctx); err != nil {
		return err
	}

	for _, conn := range app.conns {
		_ = conn.Close()
	}

	return nil
}
