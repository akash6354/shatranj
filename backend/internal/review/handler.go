package review

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

func (h *Handler) Request(w http.ResponseWriter, r *http.Request) {
	review, err := h.service.Request(r.Context(), r.PathValue("gameID"), reviewPrincipalID(r))
	if err != nil {
		writeReviewError(w, err)
		return
	}
	status := http.StatusAccepted
	if review.Status == StatusUnavailable {
		status = http.StatusServiceUnavailable
		review.Message = "Stockfish analysis is not configured"
	} else if review.Status == StatusCompleted {
		status = http.StatusOK
	}
	writeReviewJSON(w, status, review)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	review, err := h.service.Get(r.Context(), r.PathValue("gameID"), reviewPrincipalID(r))
	if err != nil {
		writeReviewError(w, err)
		return
	}
	writeReviewJSON(w, http.StatusOK, review)
}

func reviewPrincipalID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}

func writeReviewJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write review response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeReviewError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "review_not_found", "game review not found")
	case errors.Is(err, ErrForbidden):
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "you are not a player in this game")
	case errors.Is(err, ErrGameNotFinished):
		httpapi.WriteError(w, http.StatusConflict, "game_not_finished", "reviews are available after a game finishes")
	default:
		slog.Error("game review operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
