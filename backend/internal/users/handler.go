package users

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Current(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	user, err := h.service.Get(r.Context(), principal.UserID)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			httpapi.WriteError(w, http.StatusNotFound, "user_not_found", "user not found")
			return
		}
		slog.ErrorContext(r.Context(), "load current user", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, user); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
