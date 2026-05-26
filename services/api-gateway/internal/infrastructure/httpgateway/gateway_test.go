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
			Role:     authv1.UserRole_USER_ROLE_TRAINER,
			Status:   authv1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		},
		Session: &authv1.SessionInfo{
			SessionToken: "token-123",
			ExpiresAt:    timestamppb.New(time.Date(2026, time.April, 19, 12, 0, 0, 0, time.UTC)),
		},
	}, nil
}

func (server authServer) GetSession(ctx context.Context, request *authv1.GetSessionRequest) (*authv1.GetSessionResponse, error) {
	return &authv1.GetSessionResponse{
		User: &authv1.AuthUser{
			UserId:   7,
			Email:    "runner@example.com",
			Username: "runner",
			Role:     authv1.UserRole_USER_ROLE_TRAINER,
			Status:   authv1.AccountStatus_ACCOUNT_STATUS_ACTIVE,
		},
		Session: &authv1.SessionInfo{SessionToken: request.GetSessionToken()},
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
			Blocks: []*contentv1.PostBlock{
				{
					PostBlockId: 1,
					Position:    0,
					Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT,
					TextContent: stringPtr("Warm-up"),
				},
				{
					PostBlockId: 2,
					Position:    1,
					Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_IMAGE,
					FileUrl:     stringPtr("https://cdn.example/run.jpg"),
				},
				{
					PostBlockId: 3,
					Position:    2,
					Kind:        contentv1.ContentBlockKind_CONTENT_BLOCK_KIND_TEXT,
					TextContent: stringPtr("Main set"),
				},
			},
		},
	}, nil
}

func (server contentServer) CreateComment(ctx context.Context, request *contentv1.CreateCommentRequest) (*contentv1.CommentResponse, error) {
	now := timestamppb.New(time.Date(2026, time.April, 18, 12, 30, 0, 0, time.UTC))

	return &contentv1.CommentResponse{
		Comment: &contentv1.Comment{
			CommentId:    51,
			PostId:       request.GetPostId(),
			AuthorUserId: request.GetAuthorUserId(),
			Body:         request.GetBody(),
			CreatedAt:    now,
			UpdatedAt:    now,
		},
	}, nil
}

func (server contentServer) ListComments(ctx context.Context, request *contentv1.ListCommentsRequest) (*contentv1.ListCommentsResponse, error) {
	now := timestamppb.New(time.Date(2026, time.April, 18, 12, 30, 0, 0, time.UTC))

	return &contentv1.ListCommentsResponse{
		Comments: []*contentv1.Comment{
			{
				CommentId:    50,
				PostId:       request.GetPostId(),
				AuthorUserId: 8,
				Body:         "Good pace",
				CreatedAt:    now,
				UpdatedAt:    now,
			},
		},
	}, nil
}

func (server contentServer) GetTrainerStatistics(ctx context.Context, request *contentv1.GetTrainerStatisticsRequest) (*contentv1.TrainerStatisticsResponse, error) {
	return &contentv1.TrainerStatisticsResponse{
		TrainerUserId:  request.GetTrainerUserId(),
		PostsCount:     12,
		DonationsCount: 4,
		TotalRevenue:   7000,
		MonthlyRevenue: 2500,
		Currency:       request.GetCurrency(),
	}, nil
}

func (server contentServer) CreateDonationPayment(ctx context.Context, request *contentv1.CreateDonationPaymentRequest) (*contentv1.PaymentResponse, error) {
	now := timestamppb.New(time.Date(2026, time.April, 18, 12, 45, 0, 0, time.UTC))
	message := request.Message

	return &contentv1.PaymentResponse{
		Payment: &contentv1.Payment{
			PaymentId:         81,
			ProviderPaymentId: "mock-pay_abc",
			Status:            "pending",
			SenderUserId:      request.GetSenderUserId(),
			RecipientUserId:   request.GetRecipientUserId(),
			AmountValue:       request.GetAmountValue(),
			Currency:          request.GetCurrency(),
			Message:           message,
			ConfirmationToken: "confirm_abc",
			ConfirmationUrl:   "/api/v1/payments/confirm/mock/confirm_abc",
			CreatedAt:         now,
			UpdatedAt:         now,
		},
	}, nil
}

func (server contentServer) ConfirmDonationPayment(ctx context.Context, request *contentv1.ConfirmDonationPaymentRequest) (*contentv1.PaymentResponse, error) {
	now := timestamppb.New(time.Date(2026, time.April, 18, 12, 50, 0, 0, time.UTC))
	message := "Thanks"

	return &contentv1.PaymentResponse{
		Payment: &contentv1.Payment{
			PaymentId:         request.GetPaymentId(),
			ProviderPaymentId: "mock-pay_abc",
			Status:            "confirmed",
			SenderUserId:      request.GetSenderUserId(),
			RecipientUserId:   1001,
			AmountValue:       1500,
			Currency:          "RUB",
			Message:           &message,
			ConfirmationToken: request.GetConfirmationToken(),
			ConfirmationUrl:   "/api/v1/payments/confirm/mock/" + request.GetConfirmationToken(),
			Donation: &contentv1.Donation{
				DonationId:      77,
				SenderUserId:    request.GetSenderUserId(),
				RecipientUserId: 1001,
				AmountValue:     1500,
				Currency:        "RUB",
				Message:         &message,
				CreatedAt:       now,
			},
			CreatedAt:   now,
			UpdatedAt:   now,
			ConfirmedAt: now,
		},
	}, nil
}

func (server contentServer) ListNotifications(ctx context.Context, request *contentv1.ListNotificationsRequest) (*contentv1.ListNotificationsResponse, error) {
	now := timestamppb.New(time.Date(2026, time.April, 18, 13, 0, 0, 0, time.UTC))
	postID := int64(11)
	commentID := int64(50)

	return &contentv1.ListNotificationsResponse{
		Notifications: []*contentv1.Notification{
			{
				NotificationId: 91,
				UserId:         request.GetUserId(),
				Type:           "comment",
				ActorUserId:    8,
				Title:          "New comment",
				Body:           "Someone commented on your post",
				IsRead:         false,
				CreatedAt:      now,
				PostId:         &postID,
				CommentId:      &commentID,
			},
		},
	}, nil
}

func (server contentServer) MarkNotificationRead(ctx context.Context, request *contentv1.MarkNotificationReadRequest) (*contentv1.NotificationResponse, error) {
	now := timestamppb.New(time.Date(2026, time.April, 18, 13, 0, 0, 0, time.UTC))
	readAt := timestamppb.New(time.Date(2026, time.April, 18, 13, 5, 0, 0, time.UTC))

	return &contentv1.NotificationResponse{
		Notification: &contentv1.Notification{
			NotificationId: request.GetNotificationId(),
			UserId:         request.GetUserId(),
			Type:           "comment",
			ActorUserId:    8,
			Title:          "New comment",
			Body:           "Someone commented on your post",
			IsRead:         true,
			CreatedAt:      now,
			ReadAt:         readAt,
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

	handler, err := httpgateway.NewMux(context.Background(), gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer, gatewayServer)
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

	csrfResponse, err := http.Get(server.URL + "/api/v1/auth/csrf")
	if err != nil {
		t.Fatalf("get csrf token: %v", err)
	}
	defer csrfResponse.Body.Close()

	if csrfResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(csrfResponse.Body)
		t.Fatalf("unexpected csrf status: %d body=%s", csrfResponse.StatusCode, string(body))
	}

	var csrfPayload struct {
		CSRFToken string `json:"csrf_token"`
	}
	if err := json.NewDecoder(csrfResponse.Body).Decode(&csrfPayload); err != nil {
		t.Fatalf("decode csrf response: %v", err)
	}
	if strings.TrimSpace(csrfPayload.CSRFToken) == "" {
		t.Fatalf("expected csrf token in response body")
	}
	if csrfHeader := csrfResponse.Header.Get("X-CSRF-Token"); csrfHeader != csrfPayload.CSRFToken {
		t.Fatalf("expected csrf header %q, got %q", csrfPayload.CSRFToken, csrfHeader)
	}
	if setCookies := csrfResponse.Header.Values("Set-Cookie"); len(setCookies) == 0 || !strings.Contains(strings.Join(setCookies, ";"), "csrf_token="+csrfPayload.CSRFToken) {
		t.Fatalf("expected csrf cookie, got %q", setCookies)
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
		Blocks []struct {
			Kind        string `json:"kind"`
			TextContent string `json:"text_content"`
			FileURL     string `json:"file_url"`
		} `json:"blocks"`
	}
	if err := json.NewDecoder(postResponse.Body).Decode(&postPayload); err != nil {
		t.Fatalf("decode post response: %v", err)
	}
	if postPayload.PostID != 11 {
		t.Fatalf("unexpected post payload: %+v", postPayload)
	}
	if len(postPayload.Blocks) != 3 ||
		postPayload.Blocks[0].Kind != "text" ||
		postPayload.Blocks[0].TextContent != "Warm-up" ||
		postPayload.Blocks[1].Kind != "image" ||
		postPayload.Blocks[1].FileURL != "https://cdn.example/run.jpg" ||
		postPayload.Blocks[2].Kind != "text" ||
		postPayload.Blocks[2].TextContent != "Main set" {
		t.Fatalf("unexpected post blocks: %+v", postPayload.Blocks)
	}

	commentsResponse, err := http.Get(server.URL + "/api/v1/posts/11/comments?limit=20")
	if err != nil {
		t.Fatalf("list comments: %v", err)
	}
	defer commentsResponse.Body.Close()

	if commentsResponse.StatusCode != http.StatusOK {
		t.Fatalf("unexpected comments status: %d", commentsResponse.StatusCode)
	}

	var commentsPayload struct {
		Comments []struct {
			CommentID    int32  `json:"comment_id"`
			PostID       int32  `json:"post_id"`
			AuthorUserID int32  `json:"author_user_id"`
			Body         string `json:"body"`
		} `json:"comments"`
	}
	if err := json.NewDecoder(commentsResponse.Body).Decode(&commentsPayload); err != nil {
		t.Fatalf("decode comments response: %v", err)
	}
	if len(commentsPayload.Comments) != 1 ||
		commentsPayload.Comments[0].CommentID != 50 ||
		commentsPayload.Comments[0].PostID != 11 ||
		commentsPayload.Comments[0].AuthorUserID != 8 ||
		commentsPayload.Comments[0].Body != "Good pace" {
		t.Fatalf("unexpected comments payload: %+v", commentsPayload)
	}

	createCommentRequest, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/api/v1/posts/11/comments",
		bytes.NewBufferString(`{"body":"Great workout"}`),
	)
	if err != nil {
		t.Fatalf("create comment request: %v", err)
	}
	createCommentRequest.Header.Set("Content-Type", "application/json")
	createCommentRequest.Header.Set("X-CSRF-Token", csrfPayload.CSRFToken)
	createCommentRequest.Header.Set("Cookie", "sid=token-123; csrf_token="+csrfPayload.CSRFToken)

	createCommentResponse, err := http.DefaultClient.Do(createCommentRequest)
	if err != nil {
		t.Fatalf("create comment: %v", err)
	}
	defer createCommentResponse.Body.Close()

	if createCommentResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createCommentResponse.Body)
		t.Fatalf("unexpected create comment status: %d body=%s", createCommentResponse.StatusCode, string(body))
	}

	var commentPayload struct {
		Comment struct {
			CommentID    int32  `json:"comment_id"`
			PostID       int32  `json:"post_id"`
			AuthorUserID int32  `json:"author_user_id"`
			Body         string `json:"body"`
		} `json:"comment"`
	}
	if err := json.NewDecoder(createCommentResponse.Body).Decode(&commentPayload); err != nil {
		t.Fatalf("decode create comment response: %v", err)
	}
	if commentPayload.Comment.CommentID != 51 ||
		commentPayload.Comment.PostID != 11 ||
		commentPayload.Comment.AuthorUserID != 7 ||
		commentPayload.Comment.Body != "Great workout" {
		t.Fatalf("unexpected create comment payload: %+v", commentPayload)
	}

	createPaymentRequest, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/api/v1/payments/donations",
		bytes.NewBufferString(`{"user_id":1001,"amount_value":1500,"currency":"RUB","message":"Thanks"}`),
	)
	if err != nil {
		t.Fatalf("create payment request: %v", err)
	}
	createPaymentRequest.Header.Set("Content-Type", "application/json")
	createPaymentRequest.Header.Set("X-CSRF-Token", csrfPayload.CSRFToken)
	createPaymentRequest.Header.Set("Cookie", "sid=token-123; csrf_token="+csrfPayload.CSRFToken)

	createPaymentResponse, err := http.DefaultClient.Do(createPaymentRequest)
	if err != nil {
		t.Fatalf("create payment: %v", err)
	}
	defer createPaymentResponse.Body.Close()

	if createPaymentResponse.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(createPaymentResponse.Body)
		t.Fatalf("unexpected create payment status: %d body=%s", createPaymentResponse.StatusCode, string(body))
	}

	var createPaymentPayload struct {
		PaymentID         int32  `json:"payment_id"`
		Status            string `json:"status"`
		ConfirmationToken string `json:"confirmation_token"`
	}
	if err := json.NewDecoder(createPaymentResponse.Body).Decode(&createPaymentPayload); err != nil {
		t.Fatalf("decode create payment response: %v", err)
	}
	if createPaymentPayload.PaymentID != 81 ||
		createPaymentPayload.Status != "pending" ||
		createPaymentPayload.ConfirmationToken != "confirm_abc" {
		t.Fatalf("unexpected create payment payload: %+v", createPaymentPayload)
	}

	confirmPaymentRequest, err := http.NewRequest(
		http.MethodPost,
		server.URL+"/api/v1/payments/81/confirm",
		bytes.NewBufferString(`{"confirmation_token":"confirm_abc"}`),
	)
	if err != nil {
		t.Fatalf("confirm payment request: %v", err)
	}
	confirmPaymentRequest.Header.Set("Content-Type", "application/json")
	confirmPaymentRequest.Header.Set("X-CSRF-Token", csrfPayload.CSRFToken)
	confirmPaymentRequest.Header.Set("Cookie", "sid=token-123; csrf_token="+csrfPayload.CSRFToken)

	confirmPaymentResponse, err := http.DefaultClient.Do(confirmPaymentRequest)
	if err != nil {
		t.Fatalf("confirm payment: %v", err)
	}
	defer confirmPaymentResponse.Body.Close()

	if confirmPaymentResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(confirmPaymentResponse.Body)
		t.Fatalf("unexpected confirm payment status: %d body=%s", confirmPaymentResponse.StatusCode, string(body))
	}

	var confirmPaymentPayload struct {
		PaymentID int32  `json:"payment_id"`
		Status    string `json:"status"`
		Donation  struct {
			DonationID int32 `json:"donation_id"`
		} `json:"donation"`
	}
	if err := json.NewDecoder(confirmPaymentResponse.Body).Decode(&confirmPaymentPayload); err != nil {
		t.Fatalf("decode confirm payment response: %v", err)
	}
	if confirmPaymentPayload.PaymentID != 81 ||
		confirmPaymentPayload.Status != "confirmed" ||
		confirmPaymentPayload.Donation.DonationID != 77 {
		t.Fatalf("unexpected confirm payment payload: %+v", confirmPaymentPayload)
	}

	statisticsRequest, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/statistics/me", nil)
	if err != nil {
		t.Fatalf("statistics request: %v", err)
	}
	statisticsRequest.Header.Set("Cookie", "sid=token-123")

	statisticsResponse, err := http.DefaultClient.Do(statisticsRequest)
	if err != nil {
		t.Fatalf("get statistics: %v", err)
	}
	defer statisticsResponse.Body.Close()

	if statisticsResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(statisticsResponse.Body)
		t.Fatalf("unexpected statistics status: %d body=%s", statisticsResponse.StatusCode, string(body))
	}

	var statisticsPayload struct {
		TrainerID      int32  `json:"trainer_id"`
		PostsCount     int32  `json:"posts_count"`
		DonationsCount int32  `json:"donations_count"`
		TotalRevenue   int32  `json:"total_revenue"`
		MonthlyRevenue int32  `json:"monthly_revenue"`
		Currency       string `json:"currency"`
	}
	if err := json.NewDecoder(statisticsResponse.Body).Decode(&statisticsPayload); err != nil {
		t.Fatalf("decode statistics response: %v", err)
	}
	if statisticsPayload.TrainerID != 7 ||
		statisticsPayload.PostsCount != 12 ||
		statisticsPayload.DonationsCount != 4 ||
		statisticsPayload.TotalRevenue != 7000 ||
		statisticsPayload.MonthlyRevenue != 2500 ||
		statisticsPayload.Currency != "RUB" {
		t.Fatalf("unexpected statistics payload: %+v", statisticsPayload)
	}

	notificationsRequest, err := http.NewRequest(http.MethodGet, server.URL+"/api/v1/notifications?limit=20", nil)
	if err != nil {
		t.Fatalf("notifications request: %v", err)
	}
	notificationsRequest.Header.Set("Cookie", "sid=token-123")

	notificationsResponse, err := http.DefaultClient.Do(notificationsRequest)
	if err != nil {
		t.Fatalf("list notifications: %v", err)
	}
	defer notificationsResponse.Body.Close()

	if notificationsResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(notificationsResponse.Body)
		t.Fatalf("unexpected notifications status: %d body=%s", notificationsResponse.StatusCode, string(body))
	}

	var notificationsPayload struct {
		Notifications []struct {
			NotificationID int32  `json:"notification_id"`
			Type           string `json:"type"`
			IsRead         bool   `json:"is_read"`
			PostID         int32  `json:"post_id"`
			CommentID      int32  `json:"comment_id"`
		} `json:"notifications"`
	}
	if err := json.NewDecoder(notificationsResponse.Body).Decode(&notificationsPayload); err != nil {
		t.Fatalf("decode notifications response: %v", err)
	}
	if len(notificationsPayload.Notifications) != 1 ||
		notificationsPayload.Notifications[0].NotificationID != 91 ||
		notificationsPayload.Notifications[0].Type != "comment" ||
		notificationsPayload.Notifications[0].IsRead ||
		notificationsPayload.Notifications[0].PostID != 11 ||
		notificationsPayload.Notifications[0].CommentID != 50 {
		t.Fatalf("unexpected notifications payload: %+v", notificationsPayload)
	}

	markReadRequest, err := http.NewRequest(http.MethodPatch, server.URL+"/api/v1/notifications/91/read", nil)
	if err != nil {
		t.Fatalf("mark notification read request: %v", err)
	}
	markReadRequest.Header.Set("Cookie", "sid=token-123")
	markReadRequest.Header.Set("X-CSRF-Token", csrfPayload.CSRFToken)
	markReadRequest.Header.Set("Cookie", "sid=token-123; csrf_token="+csrfPayload.CSRFToken)

	markReadResponse, err := http.DefaultClient.Do(markReadRequest)
	if err != nil {
		t.Fatalf("mark notification read: %v", err)
	}
	defer markReadResponse.Body.Close()

	if markReadResponse.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(markReadResponse.Body)
		t.Fatalf("unexpected mark notification read status: %d body=%s", markReadResponse.StatusCode, string(body))
	}

	var markReadPayload struct {
		Notification struct {
			NotificationID int32 `json:"notification_id"`
			IsRead         bool  `json:"is_read"`
		} `json:"notification"`
	}
	if err := json.NewDecoder(markReadResponse.Body).Decode(&markReadPayload); err != nil {
		t.Fatalf("decode mark notification read response: %v", err)
	}
	if markReadPayload.Notification.NotificationID != 91 || !markReadPayload.Notification.IsRead {
		t.Fatalf("unexpected mark notification read payload: %+v", markReadPayload)
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

func stringPtr(value string) *string {
	return &value
}
