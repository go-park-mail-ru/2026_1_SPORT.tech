package handler

import nethttp "net/http"

type Deps struct {
	SportTypeService sportTypeService
	SessionService   sessionService
	UserService      userService
	PostService      postService
	AuthCookieName   string
}

type Handler struct {
	sportTypeService sportTypeService
	sessionService   sessionService
	userService      userService
	postService      postService
	authCookieName   string
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		sportTypeService: deps.SportTypeService,
		sessionService:   deps.SessionService,
		userService:      deps.UserService,
		postService:      deps.PostService,
		authCookieName:   deps.AuthCookieName,
	}
}

func (handler *Handler) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("GET /health", handler.handleHealth)
	mux.HandleFunc("GET /sport-types", handler.handleGetSportTypes)
	mux.HandleFunc("GET /profiles/{user_id}", handler.handleGetProfile)
	mux.HandleFunc("GET /profiles/{user_id}/posts", handler.handleGetProfilePosts)
	mux.HandleFunc("GET /posts/{post_id}", handler.handleGetPost)
	mux.HandleFunc("POST /auth/register/client", handler.handlePostAuthRegisterClient)
	mux.HandleFunc("POST /auth/register/trainer", handler.handlePostAuthRegisterTrainer)
	mux.HandleFunc("POST /auth/login", handler.handlePostAuthLogin)
	mux.Handle("GET /auth/me", handler.AuthMiddleware(nethttp.HandlerFunc(handler.handleGetAuthMe)))
	mux.Handle("POST /auth/logout", handler.AuthMiddleware(nethttp.HandlerFunc(handler.handlePostAuthLogout)))

	return mux
}

func (handler *Handler) handleHealth(writer nethttp.ResponseWriter, request *nethttp.Request) {
	writeJSON(writer, nethttp.StatusOK, healthResponse{
		Status: "ok",
	})
}
