package payments

import (
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

func (h *Handler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	principal, _ := middleware.PrincipalFromContext(r.Context())
	payment, keyID, err := h.service.CreatePremiumOrder(r.Context(), principal.UserID)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusCreated, map[string]any{
		"payment": payment, "key_id": keyID,
	}); err != nil {
		slog.Error("write payment order response", "error", err)
	}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := paymentQueryInt(r, "limit", 50)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer")
		return
	}
	offset, err := paymentQueryInt(r, "offset", 0)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "offset must be an integer")
		return
	}
	principal, _ := middleware.PrincipalFromContext(r.Context())
	items, err := h.service.List(r.Context(), principal.UserID, limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	if err := httpapi.WriteJSON(w, http.StatusOK, items); err != nil {
		slog.Error("write payment history response", "error", err)
	}
}

func (h *Handler) Webhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, 1<<20))
	if err != nil {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "webhook body exceeds 1 MiB")
		return
	}
	if err := h.service.HandleWebhook(r.Context(), body, r.Header.Get("X-Razorpay-Signature")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func paymentQueryInt(r *http.Request, key string, fallback int) (int, error) {
	if !r.URL.Query().Has(key) {
		return fallback, nil
	}
	return strconv.Atoi(r.URL.Query().Get(key))
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid payment request")
	case errors.Is(err, ErrSignatureInvalid):
		httpapi.WriteError(w, http.StatusUnauthorized, "invalid_signature", "payment webhook signature is invalid")
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "payment order not found")
	case errors.Is(err, ErrProviderUnavailable):
		httpapi.WriteError(w, http.StatusServiceUnavailable, "payments_unavailable", "payment provider is not configured")
	default:
		slog.Error("payment operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
