package notifications

import (
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	unread := r.URL.Query().Get("unread")
	if unread != "" && unread != "true" && unread != "false" {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "unread must be true or false")
		return
	}
	limit, err := queryInt(r, "limit", 50)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer")
		return
	}
	offset, err := queryInt(r, "offset", 0)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "offset must be an integer")
		return
	}
	items, err := h.service.List(r.Context(), notificationUserID(r), unread == "true", limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) UnreadCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.service.UnreadCount(r.Context(), notificationUserID(r))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"unread_count": count})
}

func (h *Handler) MarkRead(w http.ResponseWriter, r *http.Request)   { h.mark(w, r, true) }
func (h *Handler) MarkUnread(w http.ResponseWriter, r *http.Request) { h.mark(w, r, false) }

func (h *Handler) mark(w http.ResponseWriter, r *http.Request, read bool) {
	item, err := h.service.SetRead(r.Context(), notificationUserID(r), r.PathValue("notificationID"), read)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func notificationUserID(r *http.Request) string {
	principal, _ := middleware.PrincipalFromContext(r.Context())
	return principal.UserID
}

func queryInt(r *http.Request, key string, fallback int) (int, error) {
	if !r.URL.Query().Has(key) {
		return fallback, nil
	}
	return strconv.Atoi(r.URL.Query().Get(key))
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write notification response", "error", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid notification request")
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "notification not found")
	default:
		slog.Error("notification operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
