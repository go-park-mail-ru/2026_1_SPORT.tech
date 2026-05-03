package grpc

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	gatewayv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/gateway/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	sessionCookieName          = "sid"
	csrfCookieName             = "csrf_token"
	csrfHeaderName             = "X-CSRF-Token"
	httpMetadataStatusCodeKey  = "x-http-status-code"
	httpMetadataSetCookieKey   = "x-http-set-cookie"
	httpMetadataClearCookieKey = "x-http-clear-cookie"
	httpMetadataHeaderKey      = "x-http-header"
)

type Principal struct {
	SessionToken string
	User         *authv1.AuthUser
	Session      *authv1.SessionInfo
}

type Server struct {
	gatewayv1.UnimplementedAuthServiceServer
	gatewayv1.UnimplementedProfileServiceServer
	gatewayv1.UnimplementedPostServiceServer
	gatewayv1.UnimplementedTierServiceServer
	gatewayv1.UnimplementedSportServiceServer
	gatewayv1.UnimplementedDonationServiceServer
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

		outgoingMD.Set(key, append([]string(nil), values...)...)
	}

	if len(outgoingMD) == 0 {
		return ctx
	}

	return metadata.NewOutgoingContext(ctx, outgoingMD)
}

func (server *Server) requireSession(ctx context.Context) (*Principal, error) {
	sessionToken := sessionTokenFromContext(ctx)
	if sessionToken == "" {
		return nil, status.Error(codes.Unauthenticated, "unauthorized")
	}

	response, err := server.authClient.GetSession(
		forwardContext(ctx),
		&authv1.GetSessionRequest{SessionToken: sessionToken},
	)
	if err != nil {
		return nil, err
	}

	return &Principal{
		SessionToken: sessionToken,
		User:         response.GetUser(),
		Session:      response.GetSession(),
	}, nil
}

func (server *Server) optionalSession(ctx context.Context) (*Principal, error) {
	sessionToken := sessionTokenFromContext(ctx)
	if sessionToken == "" {
		return nil, nil
	}

	principal, err := server.requireSession(ctx)
	if err != nil {
		if isIgnorablePublicSessionError(err) {
			return nil, nil
		}
		return nil, err
	}

	return principal, nil
}

func (server *Server) getProfile(ctx context.Context, userID int64) (*profilev1.Profile, error) {
	response, err := server.profileClient.GetProfile(
		forwardContext(ctx),
		&profilev1.GetProfileRequest{UserId: userID},
	)
	if err != nil {
		return nil, err
	}

	return response.GetProfile(), nil
}

func sessionTokenFromContext(ctx context.Context) string {
	incomingMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := incomingMD.Get("x-session-token")
	if len(values) == 0 {
		return ""
	}

	return values[0]
}

func subscriptionLevelFromContext(ctx context.Context) (*int32, error) {
	incomingMD, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, nil
	}

	values := incomingMD.Get("x-subscription-level")
	if len(values) == 0 || values[0] == "" {
		return nil, nil
	}

	parsedValue, err := strconv.ParseInt(values[0], 10, 32)
	if err != nil || parsedValue < 1 {
		return nil, status.Error(codes.InvalidArgument, "invalid subscription level")
	}

	level := int32(parsedValue)
	return &level, nil
}

func setHTTPStatus(ctx context.Context, httpStatusCode int) error {
	return grpc.SetHeader(ctx, metadata.Pairs(httpMetadataStatusCodeKey, strconv.Itoa(httpStatusCode)))
}

func setSessionCookie(ctx context.Context, sessionToken string, expiresAt *timestamppb.Timestamp) error {
	if sessionToken == "" {
		return nil
	}

	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	if expiresAt != nil {
		cookie.Expires = expiresAt.AsTime().UTC()
	}

	return grpc.SetHeader(ctx, metadata.Pairs(httpMetadataSetCookieKey, cookie.String()))
}

func clearSessionCookie(ctx context.Context) error {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	}

	return grpc.SetHeader(ctx, metadata.Pairs(httpMetadataClearCookieKey, cookie.String()))
}

func setCSRFCookie(ctx context.Context, csrfToken string, expiresAt *timestamppb.Timestamp) error {
	if csrfToken == "" {
		return nil
	}

	cookie := &http.Cookie{
		Name:     csrfCookieName,
		Value:    csrfToken,
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
	}
	if expiresAt != nil {
		cookie.Expires = expiresAt.AsTime().UTC()
	}

	return grpc.SetHeader(ctx, metadata.Pairs(httpMetadataSetCookieKey, cookie.String()))
}

func setCSRFHeader(ctx context.Context, csrfToken string) error {
	if csrfToken == "" {
		return nil
	}

	return grpc.SetHeader(ctx, metadata.Pairs(httpMetadataHeaderKey, csrfHeaderName+":"+csrfToken))
}

func issueCSRFToken(ctx context.Context, expiresAt *timestamppb.Timestamp) (string, error) {
	csrfToken, err := newCSRFToken()
	if err != nil {
		return "", status.Errorf(codes.Internal, "generate csrf token: %v", err)
	}
	if err := setCSRFCookie(ctx, csrfToken, expiresAt); err != nil {
		return "", status.Errorf(codes.Internal, "set csrf cookie: %v", err)
	}
	if err := setCSRFHeader(ctx, csrfToken); err != nil {
		return "", status.Errorf(codes.Internal, "set csrf header: %v", err)
	}

	return csrfToken, nil
}

func clearCSRFCookie(ctx context.Context) error {
	cookie := &http.Cookie{
		Name:     csrfCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
		Expires:  time.Unix(0, 0).UTC(),
	}

	return grpc.SetHeader(ctx, metadata.Pairs(httpMetadataClearCookieKey, cookie.String()))
}

func newCSRFToken() (string, error) {
	buffer := make([]byte, 32)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(buffer), nil
}

func isIgnorablePublicSessionError(err error) bool {
	st, ok := status.FromError(err)
	if !ok {
		return false
	}

	switch st.Code() {
	case codes.Unauthenticated, codes.NotFound, codes.PermissionDenied:
		return true
	default:
		return false
	}
}

func userIDFromPrincipal(principal *Principal) (int64, error) {
	if principal == nil || principal.User == nil {
		return 0, fmt.Errorf("principal is required")
	}

	return principal.User.GetUserId(), nil
}
