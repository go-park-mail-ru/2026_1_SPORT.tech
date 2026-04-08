package handler

import (
	"log/slog"
	"net/http"
)

type Deps struct {
	Logger           *slog.Logger
	SportTypeUseCase sportTypeUseCase
	SessionUseCase   sessionUseCase
	UserUseCase      userUseCase
	PostUseCase      postUseCase
	DonationUseCase  donationUseCase
	AuthCookieName   string
}

type Handler struct {
	logger           *slog.Logger
	sportTypeUseCase sportTypeUseCase
	sessionUseCase   sessionUseCase
	userUseCase      userUseCase
	postUseCase      postUseCase
	donationUseCase  donationUseCase
	authCookieName   string
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		logger:           deps.Logger,
		sportTypeUseCase: deps.SportTypeUseCase,
		sessionUseCase:   deps.SessionUseCase,
		userUseCase:      deps.UserUseCase,
		postUseCase:      deps.PostUseCase,
		donationUseCase:  deps.DonationUseCase,
		authCookieName:   deps.AuthCookieName,
	}
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /docs", handler.handleGetDocsRedirect)
	mux.HandleFunc("GET /docs/", handler.handleGetDocs)
	mux.HandleFunc("GET /docs/openapi.yml", handler.handleGetOpenAPISpec)
	mux.HandleFunc("GET /health", handler.handleHealth)
	mux.HandleFunc("GET /sport-types", handler.handleGetSportTypes)
	mux.HandleFunc("GET /profiles/{user_id}", handler.handleGetProfile)
	mux.HandleFunc("GET /profiles/{user_id}/posts", handler.handleGetProfilePosts)
	mux.Handle("POST /profiles/{user_id}/donations", handler.AuthMiddleware(http.HandlerFunc(handler.handlePostProfileDonation)))
	mux.HandleFunc("GET /posts/{post_id}", handler.handleGetPost)
	mux.HandleFunc("POST /auth/register/client", handler.handlePostAuthRegisterClient)
	mux.HandleFunc("POST /auth/register/trainer", handler.handlePostAuthRegisterTrainer)
	mux.HandleFunc("POST /auth/login", handler.handlePostAuthLogin)
	mux.Handle("GET /auth/me", handler.AuthMiddleware(http.HandlerFunc(handler.handleGetAuthMe)))
	mux.Handle("POST /auth/logout", handler.AuthMiddleware(http.HandlerFunc(handler.handlePostAuthLogout)))

	handlerWithCORS := handler.corsMiddleware(mux)
	handlerWithRequest := handler.requestMiddleware(handlerWithCORS)

	return handlerWithRequest
}

func (handler *Handler) handleHealth(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, healthResponse{
		Status: "ok",
	})
}
