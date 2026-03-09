package handler

import (
	"context"
	nethttp "net/http"
	"regexp"

	"github.com/go-park-mail-ru/2026_1_SPORT.tech/internal/service"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,30}$`)
var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

type sessionService interface {
	CreateSession(ctx context.Context, userID int64) (string, error)
	GetUserIDBySessionID(ctx context.Context, sessionID string) (int64, error)
	RevokeSession(ctx context.Context, sessionID string) error
}

type userService interface {
	GetByID(ctx context.Context, userID int64) (service.User, error)
	RegisterClient(ctx context.Context, params service.RegisterClientParams) (service.User, error)
	Authenticate(ctx context.Context, email string, password string) (service.User, error)
}

type Deps struct {
	SessionService sessionService
	UserService    userService
	AuthCookieName string
}

type Handler struct {
	sessionService sessionService
	userService    userService
	authCookieName string
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		sessionService: deps.SessionService,
		userService:    deps.UserService,
		authCookieName: deps.AuthCookieName,
	}
}

func (handler *Handler) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("GET /health", handler.handleHealth)
	mux.HandleFunc("POST /auth/register/client", handler.handlePostAuthRegisterClient)
	mux.HandleFunc("POST /auth/login", handler.handlePostAuthLogin)
	mux.Handle("GET /auth/me", handler.AuthMiddleware(nethttp.HandlerFunc(handler.handleGetAuthMe)))

	return mux
}

func (handler *Handler) handleHealth(writer nethttp.ResponseWriter, request *nethttp.Request) {
	writeJSON(writer, nethttp.StatusOK, healthResponse{
		Status: "ok",
	})
}
