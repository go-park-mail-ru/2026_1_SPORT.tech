package handler

import (
	"context"
	nethttp "net/http"
)

type sessionService interface {
	CreateSession(ctx context.Context, userID int64) (string, error)
	GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type Deps struct {
	SessionService sessionService
	AuthCookieName string
}

type Handler struct {
	sessionService sessionService
	authCookieName string
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		sessionService: deps.SessionService,
		authCookieName: deps.AuthCookieName,
	}
}

func (handler *Handler) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("GET /health", handler.handleHealth)

	return mux
}

func (handler *Handler) handleHealth(writer nethttp.ResponseWriter, request *nethttp.Request) {
	writeJSON(writer, nethttp.StatusOK, healthResponse{
		Status: "ok",
	})
}
