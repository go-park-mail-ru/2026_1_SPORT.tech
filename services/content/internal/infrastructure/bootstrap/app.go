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

	minioadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/client/minio"
	stripeadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/client/stripe"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/grpc"
	postgresadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/adapters/repository/postgres"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/config"
	dbinfra "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/db"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/grpcserver"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/health"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/httpgateway"
	loggerinfra "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/logger"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/metrics"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/usecase"
	"google.golang.org/grpc"
)

type paymentSweeper interface {
	SweepPayments(ctx context.Context, pendingTTL time.Duration, subscriptionGrace time.Duration) (usecase.SweepResult, error)
}

type App struct {
	cfg               config.Config
	logger            *slog.Logger
	database          *sql.DB
	grpcServer        *grpc.Server
	httpServer        *http.Server
	grpcListener      net.Listener
	sweeper           paymentSweeper
	sweepInterval     time.Duration
	pendingTTL        time.Duration
	subscriptionGrace time.Duration
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	logger := loggerinfra.New(cfg.ServiceName)

	database, err := dbinfra.NewPostgres(cfg.Postgres)
	if err != nil {
		return nil, err
	}

	contentRepository := postgresadapter.NewRepository(database)
	postMediaStorage, err := minioadapter.NewPostMediaStorage(cfg.Storage)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("new post media storage: %w", err)
	}
	paymentProvider, err := stripeadapter.NewPaymentProvider(cfg.Payment)
	if err != nil {
		_ = database.Close()
		return nil, fmt.Errorf("new payment provider: %w", err)
	}
	contentUseCase := usecase.NewService(usecase.Repositories{
		Posts:                   contentRepository,
		Money:                   contentRepository,
		Engagement:              contentRepository,
		Notifications:           contentRepository,
		NotificationPreferences: contentRepository,
		Chat:                    contentRepository,
		Meeting:                 contentRepository,
	}, postMediaStorage, paymentProvider)

	metricsSet := metrics.New(cfg.ServiceName)
	grpcHandler := grpcadapter.NewServer(grpcadapter.UseCases{
		Posts:         contentUseCase,
		PostMedia:     contentUseCase,
		Tiers:         contentUseCase,
		Subscriptions: contentUseCase,
		Comments:      contentUseCase,
		Donations:     contentUseCase,
		Notifications: contentUseCase,
		Chat:          contentUseCase,
		Meeting:       contentUseCase,
	}, logger)
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

	sweepInterval, err := cfg.Payment.SweepIntervalDuration()
	if err != nil {
		_ = grpcListener.Close()
		_ = database.Close()
		return nil, fmt.Errorf("parse sweep interval: %w", err)
	}
	pendingTTL, err := cfg.Payment.PendingTTLDuration()
	if err != nil {
		_ = grpcListener.Close()
		_ = database.Close()
		return nil, fmt.Errorf("parse pending ttl: %w", err)
	}
	subscriptionGrace, err := cfg.Payment.SubscriptionGraceDuration()
	if err != nil {
		_ = grpcListener.Close()
		_ = database.Close()
		return nil, fmt.Errorf("parse subscription grace: %w", err)
	}

	stripeWebhook := stripeadapter.NewWebhookHandler(cfg.Payment.StripeWebhookSecret, contentUseCase, logger)

	httpMux := http.NewServeMux()
	httpMux.Handle("/metrics", metricsSet.Handler())
	httpMux.Handle("/healthz", health.NewHandler(cfg.ServiceName, database))
	httpMux.Handle("/openapi/content.swagger.json", httpgateway.OpenAPIHandler(cfg.OpenAPI.FilePath))
	httpMux.Handle("/webhooks/stripe", stripeWebhook)
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
		cfg:               cfg,
		logger:            logger,
		database:          database,
		grpcServer:        grpcServer,
		httpServer:        httpServer,
		grpcListener:      grpcListener,
		sweeper:           contentUseCase,
		sweepInterval:     sweepInterval,
		pendingTTL:        pendingTTL,
		subscriptionGrace: subscriptionGrace,
	}, nil
}

func (app *App) Run(ctx context.Context) error {
	errCh := make(chan error, 2)

	sweepCtx, cancelSweep := context.WithCancel(ctx)
	defer cancelSweep()
	go app.runPaymentSweeper(sweepCtx)

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

func (app *App) runPaymentSweeper(ctx context.Context) {
	ticker := time.NewTicker(app.sweepInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			result, err := app.sweeper.SweepPayments(ctx, app.pendingTTL, app.subscriptionGrace)
			if err != nil {
				app.logger.Error("payment sweep", "err", err)
				continue
			}
			if result.Confirmed > 0 || result.Expired > 0 || result.Errors > 0 || result.SubscriptionsDeactivated > 0 {
				app.logger.Info(
					"payment sweep",
					"pending_checked", result.PendingChecked,
					"confirmed", result.Confirmed,
					"expired", result.Expired,
					"errors", result.Errors,
					"subscriptions_deactivated", result.SubscriptionsDeactivated,
				)
			}
		}
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
