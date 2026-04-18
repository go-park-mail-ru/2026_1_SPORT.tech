package httpgateway_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	profilev1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/profile/v1"
	grpcadapter "github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/adapters/grpc"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type authServer struct {
	authv1.UnimplementedAuthServiceServer
}

func (server authServer) Login(ctx context.Context, request *authv1.LoginRequest) (*authv1.AuthSessionResponse, error) {
	return &authv1.AuthSessionResponse{
		User: &authv1.AuthUser{
			UserId:   7,
			Email:    request.GetEmail(),
			Username: "runner",
			Role:     authv1.UserRole_USER_ROLE_CLIENT,
			Status:   authv1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		},
		Session: &authv1.SessionInfo{
			SessionToken: "token-123",
			ExpiresAt:    timestamppb.New(time.Date(2026, time.April, 19, 12, 0, 0, 0, time.UTC)),
		},
	}, nil
}

type profileServer struct {
	profilev1.UnimplementedProfileServiceServer
}

func (server profileServer) GetProfile(ctx context.Context, request *profilev1.GetProfileRequest) (*profilev1.ProfileResponse, error) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)

	return &profilev1.ProfileResponse{
		Profile: &profilev1.Profile{
			UserId:    request.GetUserId(),
			Username:  "runner",
			FirstName: "Run",
			LastName:  "Ner",
			IsTrainer: false,
			CreatedAt: timestamppb.New(now),
			UpdatedAt: timestamppb.New(now),
		},
	}, nil
}

func (server profileServer) ListSportTypes(context.Context, *emptypb.Empty) (*profilev1.ListSportTypesResponse, error) {
	return &profilev1.ListSportTypesResponse{
		SportTypes: []*profilev1.SportType{{SportTypeId: 1, Name: "Run"}},
	}, nil
}

type contentServer struct {
	contentv1.UnimplementedContentServiceServer
}

func (server contentServer) GetPost(ctx context.Context, request *contentv1.GetPostRequest) (*contentv1.PostResponse, error) {
	now := time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)

	return &contentv1.PostResponse{
		Post: &contentv1.Post{
			PostId:       request.GetPostId(),
			AuthorUserId: 7,
			Title:        "Morning run",
			CreatedAt:    timestamppb.New(now),
			UpdatedAt:    timestamppb.New(now),
			CanView:      true,
		},
	}, nil
}

func TestNewMuxRoutesRequestsThroughGatewayFacade(t *testing.T) {
	authEndpoint := startGRPCServer(t, func(server *grpc.Server) {
		authv1.RegisterAuthServiceServer(server, authServer{})
	})
	profileEndpoint := startGRPCServer(t, func(server *grpc.Server) {
		profilev1.RegisterProfileServiceServer(server, profileServer{})
	})
	contentEndpoint := startGRPCServer(t, func(server *grpc.Server) {
		contentv1.RegisterContentServiceServer(server, contentServer{})
	})

	authConn, err := grpc.DialContext(context.Background(), authEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial auth: %v", err)
	}
	defer authConn.Close()

	profileConn, err := grpc.DialContext(context.Background(), profileEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial profile: %v", err)
	}
	defer profileConn.Close()

	contentConn, err := grpc.DialContext(context.Background(), contentEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial content: %v", err)
	}
	defer contentConn.Close()

	gatewayServer := grpcadapter.NewServer(
		authv1.NewAuthServiceClient(authConn),
		profilev1.NewProfileServiceClient(profileConn),
		contentv1.NewContentServiceClient(contentConn),
	)

	handler, err := httpgateway.NewMux(context.Background(), gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer)
	if err != nil {
		t.Fatalf("new mux: %v", err)
	}

	rootMux := http.NewServeMux()
	rootMux.Handle("/api/", http.StripPrefix("/api", handler))

	server := httptest.NewServer(rootMux)
	defer server.Close()

	loginResponse, err := http.Post(
		server.URL+"/api/v1/auth/login",
		"application/json",
		bytes.NewBufferString(`{"email":"runner@example.com","password":"secret"}`),
	)
	if err != nil {
		t.Fatalf("login request: %v", err)
	}
	defer loginResponse.Body.Close()

	if loginResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(loginResponse.Body)
		t.Fatalf("unexpected login status: %d body=%s", loginResponse.StatusCode, string(body))
	}

	var loginPayload struct {
		User struct {
			UserID    int32  `json:"user_id"`
			Username  string `json:"username"`
			Email     string `json:"email"`
			FirstName string `json:"first_name"`
			LastName  string `json:"last_name"`
		} `json:"user"`
	}
	if err := json.NewDecoder(loginResponse.Body).Decode(&loginPayload); err != nil {
		t.Fatalf("decode login response: %v", err)
	}
	if loginPayload.User.UserID != 7 ||
		loginPayload.User.Username != "runner" ||
		loginPayload.User.Email != "runner@example.com" ||
		loginPayload.User.FirstName != "Run" ||
		loginPayload.User.LastName != "Ner" {
		t.Fatalf("unexpected login payload: %+v", loginPayload)
	}
	if setCookie := loginResponse.Header.Get("Set-Cookie"); !strings.Contains(setCookie, "sid=token-123") {
		t.Fatalf("expected sid cookie, got %q", setCookie)
	}

	profileResponse, err := http.Get(server.URL + "/api/v1/profiles/7")
	if err != nil {
		t.Fatalf("get profile: %v", err)
	}
	defer profileResponse.Body.Close()

	if profileResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected profile status: %d", profileResponse.StatusCode)
	}

	var profilePayload struct {
		UserID int32 `json:"user_id"`
	}
	if err := json.NewDecoder(profileResponse.Body).Decode(&profilePayload); err != nil {
		t.Fatalf("decode profile response: %v", err)
	}
	if profilePayload.UserID != 7 {
		t.Fatalf("unexpected profile payload: %+v", profilePayload)
	}

	postResponse, err := http.Get(server.URL + "/api/v1/posts/11")
	if err != nil {
		t.Fatalf("get post: %v", err)
	}
	defer postResponse.Body.Close()

	if postResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected post status: %d", postResponse.StatusCode)
	}

	var postPayload struct {
		PostID int32 `json:"post_id"`
	}
	if err := json.NewDecoder(postResponse.Body).Decode(&postPayload); err != nil {
		t.Fatalf("decode post response: %v", err)
	}
	if postPayload.PostID != 11 {
		t.Fatalf("unexpected post payload: %+v", postPayload)
	}

	sportTypesResponse, err := http.Get(server.URL + "/api/v1/sport-types")
	if err != nil {
		t.Fatalf("get sport types: %v", err)
	}
	defer sportTypesResponse.Body.Close()

	if sportTypesResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected sport types status: %d", sportTypesResponse.StatusCode)
	}
}

func startGRPCServer(t *testing.T, register func(*grpc.Server)) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}

	server := grpc.NewServer()
	register(server)

	go func() {
		_ = server.Serve(listener)
	}()

	t.Cleanup(func() {
		server.Stop()
		_ = listener.Close()
	})

	return listener.Addr().String()
}
