package httpgateway_test

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server contentServer) ListChatConversations(_ context.Context, request *contentv1.ListChatConversationsRequest) (*contentv1.ListChatConversationsResponse, error) {
	created := timestamppb.New(time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC))
	return &contentv1.ListChatConversationsResponse{
		Conversations: []*contentv1.ChatConversation{
			{
				OtherUserId: 11,
				UnreadCount: 2,
				LastMessage: &contentv1.ChatMessage{
					MessageId:      101,
					SenderUserId:   11,
					ReceiverUserId: request.GetUserId(),
					Body:           "hi",
					IsRead:         false,
					CreatedAt:      created,
				},
			},
			{
				OtherUserId: 12,
				UnreadCount: 3,
			},
		},
	}, nil
}

type chatSnapshotDTO struct {
	Conversations []struct {
		OtherUserID int64 `json:"other_user_id"`
		UnreadCount int32 `json:"unread_count"`
		LastMessage *struct {
			MessageID int64  `json:"message_id"`
			Body      string `json:"body"`
		} `json:"last_message"`
	} `json:"conversations"`
	UnreadTotal int32 `json:"unread_total"`
}

func newChatConversationsDeps(t *testing.T) httpgateway.SSEChatDeps {
	t.Helper()

	authEndpoint := startGRPCServer(t, func(server *grpc.Server) {
		authv1.RegisterAuthServiceServer(server, authServer{})
	})
	contentEndpoint := startGRPCServer(t, func(server *grpc.Server) {
		contentv1.RegisterContentServiceServer(server, contentServer{})
	})

	authConn, err := grpc.DialContext(context.Background(), authEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial auth: %v", err)
	}
	t.Cleanup(func() { _ = authConn.Close() })

	contentConn, err := grpc.DialContext(context.Background(), contentEndpoint, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("dial content: %v", err)
	}
	t.Cleanup(func() { _ = contentConn.Close() })

	return httpgateway.SSEChatDeps{
		AuthClient:    authv1.NewAuthServiceClient(authConn),
		ContentClient: contentv1.NewContentServiceClient(contentConn),
	}
}

func readFirstSSEData(t *testing.T, body io.Reader) chatSnapshotDTO {
	t.Helper()

	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		var dto chatSnapshotDTO
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &dto); err != nil {
			t.Fatalf("decode snapshot: %v", err)
		}
		return dto
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan stream: %v", err)
	}
	t.Fatalf("no data frame received")
	return chatSnapshotDTO{}
}

func TestSSEChatConversationsHandlerStreamsSnapshot(t *testing.T) {
	deps := newChatConversationsDeps(t)
	fallback := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	server := httptest.NewServer(httpgateway.SSEChatConversationsHandler(fallback, deps))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.AddCookie(&http.Cookie{Name: "sid", Value: "token-123"})

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status: %d", resp.StatusCode)
	}
	if got := resp.Header.Get("Content-Type"); got != "text/event-stream" {
		t.Fatalf("unexpected content-type: %q", got)
	}

	snapshot := readFirstSSEData(t, resp.Body)
	cancel()

	if snapshot.UnreadTotal != 5 {
		t.Fatalf("expected unread_total 5, got %d", snapshot.UnreadTotal)
	}
	if len(snapshot.Conversations) != 2 {
		t.Fatalf("expected 2 conversations, got %d", len(snapshot.Conversations))
	}
	first := snapshot.Conversations[0]
	if first.OtherUserID != 11 || first.UnreadCount != 2 || first.LastMessage == nil || first.LastMessage.Body != "hi" {
		t.Fatalf("unexpected first conversation: %+v", first)
	}
	if snapshot.Conversations[1].LastMessage != nil {
		t.Fatalf("expected nil last_message for second conversation")
	}
}

func TestSSEChatConversationsHandlerUnauthorized(t *testing.T) {
	deps := newChatConversationsDeps(t)
	server := httptest.NewServer(httpgateway.SSEChatConversationsHandler(nil, deps))
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.StatusCode)
	}
}

func TestSSEChatConversationsHandlerFallsBackForNonGet(t *testing.T) {
	deps := newChatConversationsDeps(t)
	fallback := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	server := httptest.NewServer(httpgateway.SSEChatConversationsHandler(fallback, deps))
	defer server.Close()

	resp, err := http.Post(server.URL, "application/json", nil)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTeapot {
		t.Fatalf("expected fallback status 418, got %d", resp.StatusCode)
	}
}
