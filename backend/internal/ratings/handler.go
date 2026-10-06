package ratings

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	record, err := h.service.Get(r.Context(), principal.UserID, Mode(strings.ToLower(r.PathValue("mode"))))
	if err != nil {
		if errors.Is(err, ErrUnsupportedMode) {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_mode", "unsupported rating mode")
			return
		}
		slog.ErrorContext(r.Context(), "load user rating", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "unable to load rating")
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, record); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
