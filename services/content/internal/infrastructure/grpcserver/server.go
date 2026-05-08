package grpcserver

import (
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/content/internal/infrastructure/metrics"
	"google.golang.org/grpc"
	grpcHealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

func New(contentService contentv1.ContentServiceServer, metricSet *metrics.Metrics) *grpc.Server {
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(metricSet.UnaryServerInterceptor()),
	)

	contentv1.RegisterContentServiceServer(server, contentService)

	healthServer := grpcHealth.NewServer()
	grpc_health_v1.RegisterHealthServer(server, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(contentv1.ContentService_ServiceDesc.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(server)
	metricSet.InitializeGRPCMetrics(server)

	return server
}
