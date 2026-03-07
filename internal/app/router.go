package app

import "net/http"

type healthReponse struct {
	Status string `json:"status"`
}

func NewRouter() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", handleHealthz)

	return mux
}

func handleHealthz(writer http.ResponseWriter, request *http.Request) {
	writeJSON(writer, http.StatusOK, healthReponse{
		Status: "ok",
	})
}
