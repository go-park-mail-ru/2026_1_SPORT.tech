package handler

import (
	"encoding/json"
	nethttp "net/http"
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

func writeJSON(writer nethttp.ResponseWriter, statusCode int, body any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(statusCode)

	if body == nil {
		return
	}

	_ = json.NewEncoder(writer).Encode(body)
}

func writeError(writer nethttp.ResponseWriter, statusCode int, code string, message string) {
	writeJSON(writer, statusCode, errorResponse{
		Error: apiError{
			Code:    code,
			Message: message,
		},
	})
}

func writeValidationError(writer nethttp.ResponseWriter, fields []validationErrorField) {
	writeJSON(writer, nethttp.StatusUnprocessableEntity, errorResponse{
		Error: apiError{
			Code:    "validation_error",
			Message: "Некорректные данные",
			Fields:  fields,
		},
	})
}

func writeBadRequest(writer nethttp.ResponseWriter) {
	writeError(writer, nethttp.StatusBadRequest, "bad_request", "Некорректный запрос")
}

func writeConflict(writer nethttp.ResponseWriter, code string, message string) {
	writeError(writer, nethttp.StatusConflict, code, message)
}

func writeUnauthorized(writer nethttp.ResponseWriter) {
	writeError(writer, nethttp.StatusUnauthorized, "unauthorized", "Не авторизован")
}

func writeInternalError(writer nethttp.ResponseWriter) {
	writeError(writer, nethttp.StatusInternalServerError, "internal_error", "Внутренняя ошибка сервера")
}
