package handler

import nethttp "net/http"

type Deps struct{}

type Handler struct{}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(_ Deps) *Handler {
	return &Handler{}
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
