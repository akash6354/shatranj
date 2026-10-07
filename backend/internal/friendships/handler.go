package friendships

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.List(r.Context(), friendUserID(r))
	if err != nil {
		writeFriendError(w, err)
		return
	}
	writeFriendJSON(w, http.StatusOK, rows)
}
func (h *Handler) Requests(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.Requests(r.Context(), friendUserID(r))
	if err != nil {
		writeFriendError(w, err)
		return
	}
	writeFriendJSON(w, http.StatusOK, rows)
}
func (h *Handler) Request(w http.ResponseWriter, r *http.Request) {
	var input struct {
		UserID string `json:"user_id"`
	}
	if err := decodeFriendJSON(w, r, &input); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds limit")
		} else {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		}
		return
	}
	if err := h.service.Request(r.Context(), friendUserID(r), input.UserID); err != nil {
		writeFriendError(w, err)
		return
	}
	writeFriendJSON(w, http.StatusCreated, map[string]string{"status": "pending"})
}

func decodeFriendJSON(w http.ResponseWriter, r *http.Request, value any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 16<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return errors.New("multiple JSON values")
	}
	return nil
}
func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Accept(r.Context(), friendUserID(r), r.PathValue("userID")); err != nil {
		writeFriendError(w, err)
		return
	}
	writeFriendJSON(w, http.StatusOK, map[string]string{"status": "accepted"})
}
func (h *Handler) Reject(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Reject(r.Context(), friendUserID(r), r.PathValue("userID")); err != nil {
		writeFriendError(w, err)
		return
	}
	writeFriendJSON(w, http.StatusOK, map[string]string{"status": "rejected"})
}
func (h *Handler) Block(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Block(r.Context(), friendUserID(r), r.PathValue("userID")); err != nil {
		writeFriendError(w, err)
		return
	}
	writeFriendJSON(w, http.StatusOK, map[string]string{"status": "blocked"})
}
func (h *Handler) Unblock(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Unblock(r.Context(), friendUserID(r), r.PathValue("userID")); err != nil {
		writeFriendError(w, err)
		return
	}
	writeFriendJSON(w, http.StatusOK, map[string]string{"status": "unblocked"})
}

func friendUserID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}

func writeFriendJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write friendship response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeFriendError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "valid distinct user IDs are required")
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "friendship_not_found", "friend request or block not found")
	case errors.Is(err, ErrConflict):
		httpapi.WriteError(w, http.StatusConflict, "friendship_conflict", "friendship is not in a compatible state")
	default:
		slog.Error("friendship operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
