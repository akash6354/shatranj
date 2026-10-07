package chat

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) Rooms(w http.ResponseWriter, r *http.Request) {
	rooms, err := h.service.ListRooms(r.Context(), chatUserID(r))
	if err != nil {
		writeChatError(w, err)
		return
	}
	writeChatJSON(w, http.StatusOK, rooms)
}
func (h *Handler) Open(w http.ResponseWriter, r *http.Request) {
	var input CreateRoomInput
	if err := decodeChatJSON(w, r, &input); err != nil {
		writeChatDecodeError(w, err)
		return
	}
	room, err := h.service.Open(r.Context(), chatUserID(r), input)
	if err != nil {
		writeChatError(w, err)
		return
	}
	writeChatJSON(w, http.StatusCreated, room)
}
func (h *Handler) Messages(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if raw := r.URL.Query().Get("limit"); raw != "" {
		var err error
		limit, err = strconv.Atoi(raw)
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer")
			return
		}
	}
	messages, err := h.service.Messages(r.Context(), r.PathValue("roomID"), chatUserID(r), limit)
	if err != nil {
		writeChatError(w, err)
		return
	}
	writeChatJSON(w, http.StatusOK, messages)
}
func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var input SendInput
	if err := decodeChatJSON(w, r, &input); err != nil {
		writeChatDecodeError(w, err)
		return
	}
	message, err := h.service.Send(r.Context(), r.PathValue("roomID"), chatUserID(r), input.Content)
	if err != nil {
		writeChatError(w, err)
		return
	}
	writeChatJSON(w, http.StatusCreated, message)
}
func (h *Handler) Hide(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Reason string `json:"reason"`
	}
	if err := decodeChatJSON(w, r, &input); err != nil {
		writeChatDecodeError(w, err)
		return
	}
	if err := h.service.Hide(r.Context(), r.PathValue("roomID"), r.PathValue("messageID"), chatUserID(r), input.Reason); err != nil {
		writeChatError(w, err)
		return
	}
	writeChatJSON(w, http.StatusOK, map[string]string{"status": "hidden"})
}

func chatUserID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}
func decodeChatJSON(w http.ResponseWriter, r *http.Request, value any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if decoder.Decode(&trailing) != io.EOF {
		return errors.New("request must contain exactly one JSON value")
	}
	return nil
}
func writeChatJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write chat response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
func writeChatDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
}
func writeChatError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "chat_not_found", "chat room or message not found")
	case errors.Is(err, ErrForbidden):
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "chat room membership required")
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "message or room details are invalid")
	case errors.Is(err, ErrModerated):
		httpapi.WriteError(w, http.StatusUnprocessableEntity, "message_rejected", "message rejected by moderation")
	default:
		slog.Error("chat operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
