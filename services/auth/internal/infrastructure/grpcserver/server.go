package grpcserver

import (
	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/auth/internal/infrastructure/metrics"
	"google.golang.org/grpc"
	grpcHealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func New(authService authv1.AuthServiceServer, metricSet *metrics.Metrics) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(metricSet.UnaryServerInterceptor()),
	)

	authv1.RegisterAuthServiceServer(server, authService)

	healthServer := grpcHealth.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(authv1.AuthService_ServiceDesc.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)

	return server
}
