package auth

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var input RegisterInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeDecodeError(w, err)
		return
	}
	result, err := h.service.Register(r.Context(), input)
	if err != nil {
		writeAuthError(r, w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if err := httpapi.WriteJSON(w, http.StatusCreated, result); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var input LoginInput
	if err := decodeJSON(w, r, &input); err != nil {
		writeDecodeError(w, err)
		return
	}
	result, err := h.service.Login(r.Context(), input)
	if err != nil {
		writeAuthError(r, w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if err := httpapi.WriteJSON(w, http.StatusOK, result); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func (h *Handler) Validate(w http.ResponseWriter, r *http.Request) {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "valid bearer token required")
		return
	}
	user, err := h.service.UserByID(r.Context(), principal.UserID)
	if err != nil {
		writeAuthError(r, w, err)
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	if err := httpapi.WriteJSON(w, http.StatusOK, user); err != nil {
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return errors.New("request body must contain one JSON value")
	}
	return nil
}

func writeAuthError(r *http.Request, w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidInput):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_input", strings.TrimPrefix(err.Error(), ErrInvalidInput.Error()+": "))
	case errors.Is(err, ErrInvalidCredentials), errors.Is(err, ErrTokenInvalid):
		httpapi.WriteError(w, http.StatusUnauthorized, "unauthorized", "invalid credentials")
	case errors.Is(err, ErrEmailTaken):
		httpapi.WriteError(w, http.StatusConflict, "email_taken", "email is already registered")
	case errors.Is(err, ErrUsernameTaken):
		httpapi.WriteError(w, http.StatusConflict, "username_taken", "username is already taken")
	case errors.Is(err, ErrUserNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "user_not_found", "user not found")
	default:
		slog.ErrorContext(r.Context(), "authentication request failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
