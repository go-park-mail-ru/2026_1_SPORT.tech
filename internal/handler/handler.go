package handler

import nethttp "net/http"

type Deps struct {
	SportTypeService sportTypeService
}

type Handler struct {
	sportTypeService sportTypeService
}

type healthResponse struct {
	Status string `json:"status"`
}

func NewHandler(deps Deps) *Handler {
	return &Handler{
		sportTypeService: deps.SportTypeService,
	}
}

func (handler *Handler) Routes() nethttp.Handler {
	mux := nethttp.NewServeMux()

	mux.HandleFunc("GET /health", handler.handleHealth)
	mux.HandleFunc("GET /sport-types", handler.handleGetSportTypes)

	return mux
}

func (handler *Handler) handleHealth(writer nethttp.ResponseWriter, request *nethttp.Request) {
	writeJSON(writer, nethttp.StatusOK, healthResponse{
		Status: "ok",
	})
}
