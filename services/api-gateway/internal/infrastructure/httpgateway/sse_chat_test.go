package httpgateway_test

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/go-park-mail-ru/2026_1_SPORT.tech/services/api-gateway/internal/infrastructure/httpgateway"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server contentServer) ListChatMessages(_ context.Context, request *contentv1.ListChatMessagesRequest) (*contentv1.ListChatMessagesResponse, error) {
	return &contentv1.ListChatMessagesResponse{
		Messages: []*contentv1.ChatMessage{
			{
				MessageId:      101,
				SenderUserId:   request.GetUserId(),
				ReceiverUserId: request.GetOtherUserId(),
				Body:           "already loaded",
				IsRead:         true,
				CreatedAt:      timestamppb.New(time.Date(2026, time.April, 18, 12, 0, 0, 0, time.UTC)),
			},
		},
	}, nil
}

func TestSSEChatHandlerStreamsReadEventForLoadedMessage(t *testing.T) {
	deps := newChatConversationsDeps(t)
	fallback := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	server := httptest.NewServer(httpgateway.SSEChatHandler(fallback, deps))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL+"/api/v1/chat/messages/11/stream?after=101", nil)
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

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if line != "event: read" {
			continue
		}
		if !scanner.Scan() {
			t.Fatal("expected read event data")
		}

		dataLine := scanner.Text()
		if !strings.HasPrefix(dataLine, "data: ") {
			t.Fatalf("expected data line, got %q", dataLine)
		}

		var payload struct {
			MessageIDs []int64 `json:"message_ids"`
		}
		if err := json.Unmarshal([]byte(strings.TrimPrefix(dataLine, "data: ")), &payload); err != nil {
			t.Fatalf("decode read payload: %v", err)
		}
		if len(payload.MessageIDs) != 1 || payload.MessageIDs[0] != 101 {
			t.Fatalf("unexpected read payload: %+v", payload)
		}
		return
	}
	if err := scanner.Err(); err != nil && ctx.Err() == nil {
		t.Fatalf("scan stream: %v", err)
	}
	t.Fatal("read event was not received")
}
