package achievements

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Catalog(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.Catalog(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) Mine(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.PrincipalFromContext(r.Context())
	items, err := h.service.List(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write achievement response", "error", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalid) {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid achievement request")
		return
	}
	slog.Error("achievement operation failed", "error", err)
	httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
}
