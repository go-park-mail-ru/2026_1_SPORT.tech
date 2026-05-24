package httpgateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	authv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/auth/v1"
	contentv1 "github.com/go-park-mail-ru/2026_1_SPORT.tech/grpc/gen/go/content/v1"
	"google.golang.org/grpc/metadata"
)

const sseChatRoutePrefix = "/api/v1/chat/messages/"

type SSEChatDeps struct {
	AuthClient    authv1.AuthServiceClient
	ContentClient contentv1.ContentServiceClient
}

type sseChatMessagePayload struct {
	MessageID      int64  `json:"message_id"`
	SenderUserID   int64  `json:"sender_user_id"`
	ReceiverUserID int64  `json:"receiver_user_id"`
	Body           string `json:"body"`
	IsRead         bool   `json:"is_read"`
	CreatedAt      string `json:"created_at"`
}

func SSEChatHandler(fallback http.Handler, deps SSEChatDeps) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || !strings.HasSuffix(r.URL.Path, "/stream") {
			fallback.ServeHTTP(w, r)
			return
		}

		// Path: /api/v1/chat/messages/{other_user_id}/stream
		inner := strings.TrimPrefix(r.URL.Path, sseChatRoutePrefix)
		parts := strings.SplitN(inner, "/", 2)
		if len(parts) != 2 || parts[1] != "stream" {
			http.NotFound(w, r)
			return
		}
		otherUserID, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil || otherUserID <= 0 {
			writePublicError(w, http.StatusBadRequest, "bad_request", "invalid user id")
			return
		}

		// After-ID: client tells us the last message_id it already has.
		// This avoids re-sending history and eliminates the race between
		// loadMessages() and EventSource connect.
		afterIDStr := r.URL.Query().Get("after")
		afterID, _ := strconv.ParseInt(afterIDStr, 10, 64)

		sessionCookie, err := r.Cookie("sid")
		if err != nil || strings.TrimSpace(sessionCookie.Value) == "" {
			writePublicError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			return
		}
		sessionToken := sessionCookie.Value

		authOutMD := metadata.Pairs("x-session-token", sessionToken)
		authCtx := metadata.NewOutgoingContext(r.Context(), authOutMD)

		authResp, err := deps.AuthClient.GetSession(
			authCtx,
			&authv1.GetSessionRequest{SessionToken: sessionToken},
		)
		if err != nil || authResp.GetUser() == nil {
			writePublicError(w, http.StatusUnauthorized, "unauthorized", "unauthorized")
			return
		}
		userID := authResp.GetUser().GetUserId()

		// Disable write deadline for this long-lived connection.
		rc := http.NewResponseController(w)
		_ = rc.SetWriteDeadline(time.Time{})

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.WriteHeader(http.StatusOK)

		fmt.Fprintf(w, ": connected\n\n")
		_ = rc.Flush()

		contentOutMD := metadata.Pairs("x-session-token", sessionToken)
		contentCtx := metadata.NewOutgoingContext(r.Context(), contentOutMD)

		lastMessageID := afterID

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
				resp, err := deps.ContentClient.ListChatMessages(
					contentCtx,
					&contentv1.ListChatMessagesRequest{
						UserId:      userID,
						OtherUserId: otherUserID,
						Limit:       50,
					},
				)
				if err != nil {
					return
				}

				for _, msg := range resp.GetMessages() {
					if msg.GetMessageId() <= lastMessageID {
						continue
					}

					createdAt := ""
					if ts := msg.GetCreatedAt(); ts != nil {
						createdAt = ts.AsTime().UTC().Format(time.RFC3339Nano)
					}

					payload := sseChatMessagePayload{
						MessageID:      msg.GetMessageId(),
						SenderUserID:   msg.GetSenderUserId(),
						ReceiverUserID: msg.GetReceiverUserId(),
						Body:           msg.GetBody(),
						IsRead:         msg.GetIsRead(),
						CreatedAt:      createdAt,
					}

					data, err := json.Marshal(payload)
					if err != nil {
						continue
					}

					fmt.Fprintf(w, "data: %s\n\n", data)
					lastMessageID = msg.GetMessageId()
				}
				_ = rc.Flush()
			}
		}
	})
}
