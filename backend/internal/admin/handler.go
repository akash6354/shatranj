package admin

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

func (h *Handler) Dashboard(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Dashboard(r.Context())
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) Users(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.service.Users(r.Context(), r.URL.Query().Get("q"), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) SetUserStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	principal, _ := middleware.PrincipalFromContext(r.Context())
	if err := h.service.SetUserStatus(r.Context(), principal.UserID, r.PathValue("userID"), input.Status); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": input.Status})
}

func (h *Handler) Payments(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.service.Payments(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) FlagPayment(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Reason string `json:"reason"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	principal, _ := middleware.PrincipalFromContext(r.Context())
	flag, err := h.service.FlagPayment(r.Context(), principal.UserID, r.PathValue("paymentID"), input.Reason)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, flag)
}

func (h *Handler) PaymentFlags(w http.ResponseWriter, r *http.Request) {
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
	items, err := h.service.PaymentFlags(r.Context(), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) ResolvePaymentFlag(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.PrincipalFromContext(r.Context())
	if err := h.service.ResolvePaymentFlag(r.Context(), principal.UserID, r.PathValue("flagID")); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resolved"})
}

func (h *Handler) SetCoachStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	if err := h.service.SetCoachStatus(r.Context(), r.PathValue("userID"), input.Status); err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": input.Status})
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
		slog.Error("write admin response", "error", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid admin request")
	case errors.Is(err, ErrForbidden):
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "administrator action forbidden")
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "admin resource not found")
	default:
		slog.Error("admin operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
