package matchmaking

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	var input Request
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
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	entry, err := h.service.Join(r.Context(), principal.UserID, input)
	if err != nil {
		writeMatchmakingError(w, err)
		return
	}
	status := http.StatusAccepted
	if entry.Status == StatusPaired {
		status = http.StatusOK
	}
	if err := httpapi.WriteJSON(w, status, entry); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	if err := h.service.Leave(r.Context(), principal.UserID); err != nil {
		writeMatchmakingError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) Status(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	entry, err := h.service.Status(r.Context(), principal.UserID)
	if err != nil {
		writeMatchmakingError(w, err)
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, entry); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeMatchmakingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrQueueNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "queue_entry_not_found", "no active matchmaking entry")
	case errors.Is(err, ErrAlreadyQueued):
		httpapi.WriteError(w, http.StatusConflict, "already_queued", "user already has an active matchmaking entry")
	case errors.Is(err, ErrMatchInProgress):
		httpapi.WriteError(w, http.StatusConflict, "match_in_progress", "leave is unavailable while the match is starting or active")
	case errors.Is(err, ErrInvalidQueue):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	default:
		slog.Error("matchmaking request failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "matchmaking failed")
	}
}
