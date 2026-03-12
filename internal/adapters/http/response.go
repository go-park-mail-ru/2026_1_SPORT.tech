package handler

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

func writeNoContent(writer http.ResponseWriter) {
	writer.WriteHeader(http.StatusNoContent)
}

func writeError(writer http.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(writer, statusCode, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func writeValidationError(writer http.ResponseWriter, fields []validationErrorField) {
	writeJSON(writer, http.StatusUnprocessableEntity, errorResponse{
		Error: apiError{
			Code:    "validation_error",
			Message: "Некорректные данные",
			Fields:  fields,
		},
	})
}

func writeBadRequest(writer http.ResponseWriter) {
	writeError(writer, http.StatusBadRequest, "bad_request", "Некорректный запрос")
}

func writeConflict(writer http.ResponseWriter, code string, message string) {
	writeError(writer, http.StatusConflict, code, message)
}

func writeUnauthorized(writer http.ResponseWriter) {
	writeError(writer, http.StatusUnauthorized, "unauthorized", "Не авторизован")
}

func writeForbidden(writer http.ResponseWriter, message string) {
	writeError(writer, http.StatusForbidden, "forbidden", message)
}

func writeNotFound(writer http.ResponseWriter, message string) {
	writeError(writer, http.StatusNotFound, "not_found", message)
}

func writeInvalidCredentials(writer http.ResponseWriter) {
	writeError(writer, http.StatusUnauthorized, "invalid_credentials", "Неверный email или пароль")
}

func writeInternalError(writer http.ResponseWriter) {
	writeError(writer, http.StatusInternalServerError, "internal_error", "Внутренняя ошибка сервера")
}
