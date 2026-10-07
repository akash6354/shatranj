package profiles

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	profile, err := h.service.GetByUsername(r.Context(), r.PathValue("username"))
	if err != nil {
		writeProfileError(w, err)
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, profile); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	profile, err := h.service.GetByUserID(r.Context(), principal.UserID)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, profile); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	var input UpdateInput
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		} else {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		}
		return
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must contain one JSON value")
		return
	}
	profile, err := h.service.Update(r.Context(), principal.UserID, input)
	if err != nil {
		writeProfileError(w, err)
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, profile); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeProfileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "profile_not_found", "profile not found")
	case errors.Is(err, ErrUsernameTaken):
		httpapi.WriteError(w, http.StatusConflict, "username_taken", "username is already taken")
	case errors.Is(err, ErrInvalidInput):
		message := strings.TrimPrefix(err.Error(), ErrInvalidInput.Error()+": ")
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_input", message)
	default:
		slog.Error("profile request failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
