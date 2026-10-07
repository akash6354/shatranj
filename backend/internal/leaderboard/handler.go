package leaderboard

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/ratings"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Top(w http.ResponseWriter, r *http.Request) {
	limit := defaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer")
			return
		}
		limit = parsed
	}
	entries, err := h.service.Top(r.Context(), ratings.Mode(r.PathValue("mode")), limit)
	if err != nil {
		writeLeaderboardError(w, err)
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, entries); err != nil {
		slog.Error("write leaderboard response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeLeaderboardError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidRequest):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	default:
		slog.Error("leaderboard request failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
