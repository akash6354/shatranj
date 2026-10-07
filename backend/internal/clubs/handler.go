package clubs

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
	clubs, err := h.service.List(r.Context(), clubUserID(r))
	if err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, clubs)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := decodeClubJSON(w, r, &input); err != nil {
		writeClubDecodeError(w, err)
		return
	}
	club, err := h.service.Create(r.Context(), clubUserID(r), input)
	if err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusCreated, club)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	club, err := h.service.Get(r.Context(), r.PathValue("clubID"), clubUserID(r))
	if err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, club)
}

func (h *Handler) Members(w http.ResponseWriter, r *http.Request) {
	members, err := h.service.Members(r.Context(), r.PathValue("clubID"), clubUserID(r))
	if err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, members)
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	status, err := h.service.Join(r.Context(), r.PathValue("clubID"), clubUserID(r))
	if err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, map[string]string{"status": status})
}

func (h *Handler) Leave(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Leave(r.Context(), r.PathValue("clubID"), clubUserID(r)); err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, map[string]string{"status": "left"})
}

func (h *Handler) Requests(w http.ResponseWriter, r *http.Request) {
	requests, err := h.service.Requests(r.Context(), r.PathValue("clubID"), clubUserID(r))
	if err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, requests)
}

func (h *Handler) ResolveRequest(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Approve bool `json:"approve"`
	}
	if err := decodeClubJSON(w, r, &input); err != nil {
		writeClubDecodeError(w, err)
		return
	}
	err := h.service.ResolveRequest(r.Context(), r.PathValue("clubID"), r.PathValue("requestID"), clubUserID(r), input.Approve)
	if err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (h *Handler) SetRole(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Role string `json:"role"`
	}
	if err := decodeClubJSON(w, r, &input); err != nil {
		writeClubDecodeError(w, err)
		return
	}
	if err := h.service.SetRole(r.Context(), r.PathValue("clubID"), r.PathValue("userID"), clubUserID(r), input.Role); err != nil {
		writeClubError(w, err)
		return
	}
	writeClubJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func clubUserID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}

func decodeClubJSON(w http.ResponseWriter, r *http.Request, value any) error {
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

func writeClubJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write club response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeClubDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
}

func writeClubError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "club_not_found", "club not found")
	case errors.Is(err, ErrForbidden):
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "club membership or moderation permission required")
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, ErrConflict):
		httpapi.WriteError(w, http.StatusConflict, "club_conflict", "club membership request conflicts with current state")
	default:
		slog.Error("club operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
