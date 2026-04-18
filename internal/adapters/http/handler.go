package handler

import (
	"log/slog"
	"net/http"
)

type Deps struct {
	Logger               *slog.Logger
	SportTypeUseCase     sportTypeUseCase
	SessionUseCase       sessionUseCase
	UserUseCase          userUseCase
	PostUseCase          postUseCase
	DonationUseCase      donationUseCase
	AuthCookieName       string
	StoragePublicBaseURL string
}

type Handler struct {
	logger               *slog.Logger
	sportTypeUseCase     sportTypeUseCase
	sessionUseCase       sessionUseCase
	userUseCase          userUseCase
	postUseCase          postUseCase
	donationUseCase      donationUseCase
	authCookieName       string
	storagePublicBaseURL string
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		logger:               deps.Logger,
		sportTypeUseCase:     deps.SportTypeUseCase,
		sessionUseCase:       deps.SessionUseCase,
		userUseCase:          deps.UserUseCase,
		postUseCase:          deps.PostUseCase,
		donationUseCase:      deps.DonationUseCase,
		authCookieName:       deps.AuthCookieName,
		storagePublicBaseURL: deps.StoragePublicBaseURL,
	}
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.handleHealth)
	mux.HandleFunc("GET /sport-types", handler.handleGetSportTypes)
	mux.HandleFunc("GET /trainers", handler.handleGetTrainers)
	mux.HandleFunc("GET /profiles/{user_id}", handler.handleGetProfile)
	mux.Handle("PATCH /profiles/me", handler.AuthMiddleware(http.HandlerFunc(handler.handlePatchProfileMe)))
	mux.Handle("POST /profiles/me/avatar", handler.AuthMiddleware(http.HandlerFunc(handler.handlePostProfileAvatar)))
	mux.Handle("DELETE /profiles/me/avatar", handler.AuthMiddleware(http.HandlerFunc(handler.handleDeleteProfileAvatar)))
	mux.HandleFunc("GET /profiles/{user_id}/posts", handler.handleGetProfilePosts)
	mux.Handle("POST /profiles/{user_id}/donations", handler.AuthMiddleware(http.HandlerFunc(handler.handlePostProfileDonation)))
	mux.HandleFunc("GET /posts/{post_id}", handler.handleGetPost)
	mux.Handle("POST /posts/{post_id}/likes", handler.AuthMiddleware(http.HandlerFunc(handler.handlePostPostLike)))
	mux.Handle("DELETE /posts/{post_id}/likes", handler.AuthMiddleware(http.HandlerFunc(handler.handleDeletePostLike)))
	mux.Handle("POST /posts", handler.AuthMiddleware(http.HandlerFunc(handler.handlePostCreate)))
	mux.Handle("PATCH /posts/{post_id}", handler.AuthMiddleware(http.HandlerFunc(handler.handlePatchPost)))
	mux.Handle("DELETE /posts/{post_id}", handler.AuthMiddleware(http.HandlerFunc(handler.handleDeletePost)))
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
