package coaches

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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.service.List(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), r.PathValue("coachID"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) Upsert(w http.ResponseWriter, r *http.Request) {
	var input ProfileInput
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	principal, _ := middleware.PrincipalFromContext(r.Context())
	item, err := h.service.Upsert(r.Context(), principal.UserID, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) Book(w http.ResponseWriter, r *http.Request) {
	var input BookingInput
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	principal, _ := middleware.PrincipalFromContext(r.Context())
	item, err := h.service.Book(r.Context(), principal.UserID, r.PathValue("coachID"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) Bookings(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.PrincipalFromContext(r.Context())
	items, err := h.service.Bookings(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) UpdateBooking(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	principal, _ := middleware.PrincipalFromContext(r.Context())
	item, err := h.service.UpdateBooking(r.Context(), principal.UserID, r.PathValue("bookingID"), input.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func queryInt(r *http.Request, key string, fallback int) (int, error) {
	if !r.URL.Query().Has(key) {
		return fallback, nil
	}
	return strconv.Atoi(r.URL.Query().Get(key))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request contains trailing data")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write coach response", "error", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid coach or booking request")
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "coach or booking not found")
	case errors.Is(err, ErrConflict):
		httpapi.WriteError(w, http.StatusConflict, "booking_conflict", "the requested time is unavailable")
	case errors.Is(err, ErrForbidden):
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "coach action is not allowed")
	default:
		slog.Error("coach operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
