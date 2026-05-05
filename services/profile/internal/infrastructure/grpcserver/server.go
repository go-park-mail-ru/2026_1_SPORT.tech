package grpcserver

import (
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/profile/internal/infrastructure/metrics"
	"google.golang.org/grpc"
	grpcHealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func New(profileService profilev1.ProfileServiceServer, metricSet *metrics.Metrics) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(metricSet.UnaryServerInterceptor()),
	)

	profilev1.RegisterProfileServiceServer(server, profileService)

	healthServer := grpcHealth.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(profilev1.ProfileService_ServiceDesc.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)
	metricSet.InitializeGRPCMetrics(server)

	return server
}
