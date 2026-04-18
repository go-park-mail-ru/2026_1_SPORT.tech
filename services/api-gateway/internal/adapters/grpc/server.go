package grpc

import (
	"context"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"google.golang.org/grpc/metadata"
)

type Server struct {
	gatewayv1.UnimplementedAuthServiceServer
	gatewayv1.UnimplementedProfileServiceServer
	gatewayv1.UnimplementedContentServiceServer
	authClient    authv1.AuthServiceClient
	profileClient profilev1.ProfileServiceClient
	contentClient contentv1.ContentServiceClient
}

func NewServer(
	authClient authv1.AuthServiceClient,
	profileClient profilev1.ProfileServiceClient,
	contentClient contentv1.ContentServiceClient,
) *Server {
	return &Server{
		authClient:    authClient,
		profileClient: profileClient,
		contentClient: contentClient,
	}
}

func forwardContext(ctx context.Context) context.Context {
	incomingMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx
	}

	outgoingMD := metadata.MD{}
	for _, key := range []string{"authorization", "x-request-id", "x-session-token", "x-user-id", "x-subscription-level"} {
		values := incomingMD.Get(key)
		if len(values) == 0 {
			continue
		}

		copied := append([]string(nil), values...)
		outgoingMD.Set(key, copied...)
	}

	if len(outgoingMD) == 0 {
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, outgoingMD)
}
