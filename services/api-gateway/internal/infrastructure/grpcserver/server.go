package grpcserver

import (
	"context"
	"fmt"
	"net"

	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/metrics"
	"google.golang.org/grpc"
	grpcHealth "google.golang.org/grpc/health"
	grpcHealthV1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	server   *grpc.Server
	listener net.Listener
}

func New(
	listenAddress string,
	authService gatewayv1.AuthServiceServer,
	profileService gatewayv1.ProfileServiceServer,
	postService gatewayv1.PostServiceServer,
	tierService gatewayv1.TierServiceServer,
	sportService gatewayv1.SportServiceServer,
	donationService gatewayv1.DonationServiceServer,
	metricSet *metrics.Metrics,
) (*Server, error) {
	listener, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return nil, fmt.Errorf("listen grpc: %w", err)
	}

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(metricSet.UnaryServerInterceptor()),
	)
	gatewayv1.RegisterAuthServiceServer(grpcServer, authService)
	gatewayv1.RegisterProfileServiceServer(grpcServer, profileService)
	gatewayv1.RegisterPostServiceServer(grpcServer, postService)
	gatewayv1.RegisterTierServiceServer(grpcServer, tierService)
	gatewayv1.RegisterSportServiceServer(grpcServer, sportService)
	gatewayv1.RegisterDonationServiceServer(grpcServer, donationService)

	healthServer := grpcHealth.NewServer()
	healthServer.SetServingStatus("", grpcHealthV1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(gatewayv1.AuthService_ServiceDesc.ServiceName, grpcHealthV1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(gatewayv1.ProfileService_ServiceDesc.ServiceName, grpcHealthV1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(gatewayv1.PostService_ServiceDesc.ServiceName, grpcHealthV1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(gatewayv1.TierService_ServiceDesc.ServiceName, grpcHealthV1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(gatewayv1.SportService_ServiceDesc.ServiceName, grpcHealthV1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus(gatewayv1.DonationService_ServiceDesc.ServiceName, grpcHealthV1.HealthCheckResponse_SERVING)
	grpcHealthV1.RegisterHealthServer(grpcServer, healthServer)

	reflection.Register(grpcServer)

	return &Server{
		server:   grpcServer,
		listener: listener,
	}, nil
}

func (server *Server) Serve() error {
	return server.server.Serve(server.listener)
}

func (server *Server) Shutdown(ctx context.Context) error {
	done := make(chan struct{})

	go func() {
		server.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		server.server.Stop()
		<-done
		return nil
	}
}

func (server *Server) Address() string {
	return server.listener.Addr().String()
}
