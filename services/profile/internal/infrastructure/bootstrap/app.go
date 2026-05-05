package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	minioadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/client/minio"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/grpc"
	postgresadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/adapters/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/config"
	dbinfra "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/db"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/grpcserver"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/health"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/httpgateway"
	loggerinfra "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/logger"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/metrics"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/usecase"
	"google.golang.org/grpc"
)

type App struct {
	cfg          config.Config
	logger       *slog.Logger
	database     *sql.DB
	grpcServer   *grpc.Server
	httpServer   *http.Server
	grpcListener net.Listener
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	logger := loggerinfra.New(cfg.ServiceName)

	database, err := dbinfra.NewPostgres(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	profileRepository := postgresadapter.NewProfileRepository(database)
	sportTypeRepository := postgresadapter.NewSportTypeRepository(database)
	avatarStorage, err := minioadapter.NewAvatarStorage(cfg.Storage)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("new avatar storage: %w", err)
	}
	profileUseCase := usecase.NewService(profileRepository, sportTypeRepository, avatarStorage)

	metricsSet := metrics.New(cfg.ServiceName)
	grpcHandler := grpcadapter.NewServer(grpcadapter.UseCases{
		Profiles: profileUseCase,
		Authors:  profileUseCase,
		Avatars:  profileUseCase,
		Sports:   profileUseCase,
	})
	grpcServer := grpcserver.New(grpcHandler, metricsSet)

	grpcListener, err := net.Listen("tcp", cfg.Server.GRPCAddress())
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("listen grpc: %w", err)
	}

	gatewayHandler, err := httpgateway.NewLocalMux(ctx, grpcHandler)
	if err != nil {
		_ = grpcListener.Close()
		_ = database.Close()
		return nil, fmt.Errorf("new local gateway: %w", err)
	}

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", metricsSet.Handler())
	httpMux.Handle("/healthz", health.NewHandler(cfg.ServiceName, database))
	httpMux.Handle("/openapi/profile.swagger.json", httpgateway.OpenAPIHandler(cfg.OpenAPI.FilePath))
	httpMux.Handle("/", gatewayHandler)

	httpServer := &http.Server{
		Addr:              cfg.Server.HTTPAddress(),
		Handler:           metricsSet.HTTPMiddleware(httpMux),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return &App{
		cfg:          cfg,
		logger:       logger,
		database:     database,
		grpcServer:   grpcServer,
		httpServer:   httpServer,
		grpcListener: grpcListener,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	go func() {
		app.logger.Info("starting gRPC server", "addr", app.cfg.Server.GRPCAddress())
		if err := app.grpcServer.Serve(app.grpcListener); err != nil && !errors.Is(err, grpc.ErrServerStopped) && !errors.Is(err, net.ErrClosed) {
			errCh <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	go func() {
		app.logger.Info("starting HTTP gateway", "addr", app.cfg.Server.HTTPAddress())
		if err := app.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- fmt.Errorf("http server: %w", err)
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

	var shutdownErr error
	if err := app.grpcListener.Close(); err != nil && !errors.Is(err, net.ErrClosed) {
		shutdownErr = errors.Join(shutdownErr, err)
	}

	done := make(chan struct{})
	go func() {
		app.grpcServer.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		app.grpcServer.Stop()
	}

	if err := app.httpServer.Shutdown(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		shutdownErr = errors.Join(shutdownErr, err)
	}
	if err := app.database.Close(); err != nil {
		shutdownErr = errors.Join(shutdownErr, err)
	}

	return shutdownErr
}
