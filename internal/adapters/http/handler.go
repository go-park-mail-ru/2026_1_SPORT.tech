package handler

import "net/http"

type Deps struct {
	SportTypeUseCase sportTypeUseCase
	SessionUseCase   sessionUseCase
	UserUseCase      userUseCase
	PostUseCase      postUseCase
	AuthCookieName   string
}

type Handler struct {
	sportTypeUseCase sportTypeUseCase
	sessionUseCase   sessionUseCase
	userUseCase      userUseCase
	postUseCase      postUseCase
	authCookieName   string
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		sportTypeUseCase: deps.SportTypeUseCase,
		sessionUseCase:   deps.SessionUseCase,
		userUseCase:      deps.UserUseCase,
		postUseCase:      deps.PostUseCase,
		authCookieName:   deps.AuthCookieName,
	}
}

func (handler *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", handler.handleHealth)
	mux.HandleFunc("GET /sport-types", handler.handleGetSportTypes)
	mux.HandleFunc("GET /profiles/{user_id}", handler.handleGetProfile)
	mux.HandleFunc("GET /profiles/{user_id}/posts", handler.handleGetProfilePosts)
	mux.HandleFunc("GET /posts/{post_id}", handler.handleGetPost)
	mux.HandleFunc("POST /auth/register/client", handler.handlePostAuthRegisterClient)
	mux.HandleFunc("POST /auth/register/trainer", handler.handlePostAuthRegisterTrainer)
	mux.HandleFunc("POST /auth/login", handler.handlePostAuthLogin)
	mux.Handle("GET /auth/me", handler.AuthMiddleware(http.HandlerFunc(handler.handleGetAuthMe)))
	mux.Handle("POST /auth/logout", handler.AuthMiddleware(http.HandlerFunc(handler.handlePostAuthLogout)))

	return handler.corsMiddleware(mux)
}

func (handler *Handler) handleHealth(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, healthResponse{
		Status: "ok",
	})
}
