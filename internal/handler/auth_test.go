package handler

import (
	"context"
	"errors"
	nethttp "net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

type sessionServiceStub struct {
	createSessionFunc        func(ctx context.Context, userID int64) (string, error)
	getUserIDBySessionIDFunc func(ctx context.Context, sessionID string) (int64, error)
	revokeSessionFunc        func(ctx context.Context, sessionID string) error
}

func (stub *sessionServiceStub) CreateSession(ctx context.Context, userID int64) (string, error) {
	if stub.createSessionFunc == nil {
		return "", nil
	}

	return stub.createSessionFunc(ctx, userID)
}

func (stub *sessionServiceStub) GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error) {
	if stub.getUserIDBySessionIDFunc == nil {
		return 0, nil
	}

	return stub.getUserIDBySessionIDFunc(ctx, sessionID)
}

func (stub *sessionServiceStub) RevokeSession(ctx context.Context, sessionID string) error {
	if stub.revokeSessionFunc == nil {
		return nil
	}

	return stub.revokeSessionFunc(ctx, sessionID)
}

type logoutHandlerTest struct {
	name            string
	sessionID       string
	sessionService  *sessionServiceStub
	expectStatus    int
	expectBody      string
	expectSetCookie bool
}

func TestHandlePostAuthLogoutPositive(t *testing.T) {
	tests := []logoutHandlerTest{
		{
			name:      "Успешный logout",
			sessionID: "valid-session-id",
			sessionService: &sessionServiceStub{
				getUserIDBySessionIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
					return 42, nil
				},
				revokeSessionFunc: func(ctx context.Context, sessionID string) error {
					if sessionID != "valid-session-id" {
						t.Fatalf("unexpected session id: got %s, expect %s", sessionID, "valid-session-id")
					}

					return nil
				},
			},
			expectStatus:    nethttp.StatusNoContent,
			expectSetCookie: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(Deps{
				SessionService: tt.sessionService,
				AuthCookieName: "sid",
			})

			request := httptest.NewRequest(nethttp.MethodPost, "/auth/logout", nil)
			request.AddCookie(&nethttp.Cookie{
				Name:  "sid",
				Value: tt.sessionID,
			})

			recorder := httptest.NewRecorder()
			handler.Routes().ServeHTTP(recorder, request)

			if recorder.Code != tt.expectStatus {
				t.Fatalf("unexpected status: got %d, expect %d", recorder.Code, tt.expectStatus)
			}

			setCookie := recorder.Header().Get("Set-Cookie")
			if tt.expectSetCookie && !strings.Contains(setCookie, "sid=") {
				t.Fatalf("expected sid cookie to be cleared, got %q", setCookie)
			}
		})
	}
}

func TestHandlePostAuthLogoutNegative(t *testing.T) {
	internalErr := errors.New("revoke session")
	tests := []logoutHandlerTest{
		{
			name:           "Нет cookie",
			sessionService: &sessionServiceStub{},
			expectStatus:   nethttp.StatusUnauthorized,
			expectBody:     `"code":"unauthorized"`,
		},
		{
			name:      "Сессия не найдена",
			sessionID: "missing-session-id",
			sessionService: &sessionServiceStub{
				getUserIDBySessionIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
					return 0, service.ErrSessionNotFound
				},
			},
			expectStatus: nethttp.StatusUnauthorized,
			expectBody:   `"code":"unauthorized"`,
		},
		{
			name:      "Ошибка revoke",
			sessionID: "valid-session-id",
			sessionService: &sessionServiceStub{
				getUserIDBySessionIDFunc: func(ctx context.Context, sessionID string) (int64, error) {
					return 42, nil
				},
				revokeSessionFunc: func(ctx context.Context, sessionID string) error {
					return internalErr
				},
			},
			expectStatus: nethttp.StatusInternalServerError,
			expectBody:   `"code":"internal_error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewHandler(Deps{
				SessionService: tt.sessionService,
				AuthCookieName: "sid",
			})

			request := httptest.NewRequest(nethttp.MethodPost, "/auth/logout", nil)
			if tt.sessionID != "" {
				request.AddCookie(&nethttp.Cookie{
					Name:  "sid",
					Value: tt.sessionID,
				})
			}

			recorder := httptest.NewRecorder()
			handler.Routes().ServeHTTP(recorder, request)

			if recorder.Code != tt.expectStatus {
				t.Fatalf("unexpected status: got %d, expect %d", recorder.Code, tt.expectStatus)
			}
			if !strings.Contains(recorder.Body.String(), tt.expectBody) {
				t.Fatalf("unexpected body: got %q, expect body to contain %q", recorder.Body.String(), tt.expectBody)
			}
		})
	}
}
