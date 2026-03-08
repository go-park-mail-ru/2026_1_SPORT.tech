package app

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string                 `json:"code"`
	Message string                 `json:"message"`
	Fields  []validationErrorField `json:"fields,omitempty"`
}

type validationErrorField struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func writeJSON(writer http.ResponseWriter, statusCode int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	if body == nil {
		return
	}

	_ = json.NewEncoder(writer).Encode(body)
}

func writeError(writer http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(writer, statusCode, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func writeBadRequest(writer http.ResponseWriter) {
	writeError(writer, http.StatusBadRequest, "bad_request", "Некорректный запрос")
}

func writeUnauthorized(writer http.ResponseWriter) {
	writeError(writer, http.StatusUnauthorized, "unauthorized", "Не авторизован")
}

func writeNotFound(writer http.ResponseWriter, message string) {
	writeError(writer, http.StatusNotFound, "not_found", message)
}
