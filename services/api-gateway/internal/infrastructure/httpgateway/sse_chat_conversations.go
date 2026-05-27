package httpgateway

//go:generate go run github.com/mailru/easyjson/easyjson $GOFILE

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"github.com/mailru/easyjson"
	"google.golang.org/grpc/metadata"
)

//easyjson:json
type sseConversationPayload struct {
	OtherUserID int64                  `json:"other_user_id"`
	LastMessage *sseChatMessagePayload `json:"last_message"`
	UnreadCount int32                  `json:"unread_count"`
}

//easyjson:json
type sseConversationsSnapshot struct {
	Conversations []sseConversationPayload `json:"conversations"`
	UnreadTotal   int32                    `json:"unread_total"`
}

func resolveSSESession(w http.ResponseWriter, r *http.Request, authClient authv1.AuthServiceClient) (int64, string, bool) {
	sessionCookie, err := r.Cookie("sid")
	if err != nil || strings.TrimSpace(sessionCookie.Value) == "" {
		writePublicError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return 0, "", false
	}
	sessionToken := sessionCookie.Value

	authCtx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs("x-session-token", sessionToken))
	authResp, err := authClient.GetSession(authCtx, &authv1.GetSessionRequest{SessionToken: sessionToken})
	if err != nil || authResp.GetUser() == nil {
		writePublicError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
		return 0, "", false
	}
	return authResp.GetUser().GetUserId(), sessionToken, true
}

func SSEChatConversationsHandler(fallback http.Handler, deps SSEChatDeps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			fallback.ServeHTTP(w, r)
			return
		}

		userID, sessionToken, ok := resolveSSESession(w, r, deps.AuthClient)
		if !ok {
			return
		}

		rc := http.NewResponseController(w)
		_ = rc.SetWriteDeadline(time.Time{})

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache, no-transform")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Accel-Buffering", "no")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, ": connected\n\n")
		_ = rc.Flush()

		contentCtx := metadata.NewOutgoingContext(r.Context(), metadata.Pairs("x-session-token", sessionToken))

		lastSignature := ""
		emit := func() bool {
			snapshot, signature, err := fetchConversationsSnapshot(contentCtx, deps.ContentClient, userID)
			if err != nil {
				return false
			}
			if signature == lastSignature {
				return true
			}
			lastSignature = signature

			data, err := easyjson.Marshal(snapshot)
			if err != nil {
				return true
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			_ = rc.Flush()
			return true
		}

		if !emit() {
			return
		}

		pollTicker := time.NewTicker(500 * time.Millisecond)
		defer pollTicker.Stop()
		keepaliveTicker := time.NewTicker(15 * time.Second)
		defer keepaliveTicker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-keepaliveTicker.C:
				fmt.Fprintf(w, ": keepalive\n\n")
				_ = rc.Flush()
			case <-pollTicker.C:
				if !emit() {
					return
				}
			}
		}
	})
}

func fetchConversationsSnapshot(ctx context.Context, client contentv1.ContentServiceClient, userID int64) (sseConversationsSnapshot, string, error) {
	resp, err := client.ListChatConversations(ctx, &contentv1.ListChatConversationsRequest{UserId: userID})
	if err != nil {
		return sseConversationsSnapshot{}, "", err
	}

	conversations := make([]sseConversationPayload, 0, len(resp.GetConversations()))
	var unreadTotal int32
	var sb strings.Builder

	for _, conv := range resp.GetConversations() {
		unreadTotal += conv.GetUnreadCount()

		var lastMessage *sseChatMessagePayload
		var lastMessageID int64
		if msg := conv.GetLastMessage(); msg != nil {
			createdAt := ""
			if ts := msg.GetCreatedAt(); ts != nil {
				createdAt = ts.AsTime().UTC().Format(time.RFC3339Nano)
			}
			lastMessageID = msg.GetMessageId()
			lastMessage = &sseChatMessagePayload{
				MessageID:      msg.GetMessageId(),
				SenderUserID:   msg.GetSenderUserId(),
				ReceiverUserID: msg.GetReceiverUserId(),
				Body:           msg.GetBody(),
				IsRead:         msg.GetIsRead(),
				CreatedAt:      createdAt,
			}
		}

		conversations = append(conversations, sseConversationPayload{
			OtherUserID: conv.GetOtherUserId(),
			LastMessage: lastMessage,
			UnreadCount: conv.GetUnreadCount(),
		})

		fmt.Fprintf(&sb, "%d:%d:%d:%t|", conv.GetOtherUserId(), lastMessageID, conv.GetUnreadCount(), conv.GetLastMessage().GetIsRead())
	}

	return sseConversationsSnapshot{Conversations: conversations, UnreadTotal: unreadTotal}, sb.String(), nil
}
