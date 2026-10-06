package httpapi

import (
	"log/slog"
	"net/http"
)

// APIError is the public error envelope returned by the API.
type APIError struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail gives clients a stable machine-readable code and message.
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteError writes an APIError using the standard JSON response format.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	if err := WriteJSON(w, status, APIError{Error: ErrorDetail{Code: code, Message: message}}); err != nil {
		slog.Error("failed to write API error", "error", err)
	}
}
