package httpapi

import (
	"encoding/json"
	"net/http"
)

type errorResponse struct {
	Error apiError `json:"error"`
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// WriteJSON writes a JSON response with the supplied HTTP status.
func WriteJSON(w http.ResponseWriter, status int, value any) error {
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	body = append(body, '\n')

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, err = w.Write(body)
	return err
}

// WriteError writes the standard API error envelope.
func WriteError(w http.ResponseWriter, status int, code, message string) error {
	return WriteJSON(w, status, errorResponse{
		Error: apiError{Code: code, Message: message},
	})
}
