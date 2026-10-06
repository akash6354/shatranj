package news

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	limit, err := newsQueryInt(r, "limit", 20)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer")
		return
	}
	offset, err := newsQueryInt(r, "offset", 0)
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

func (h *Handler) AdminList(w http.ResponseWriter, r *http.Request) {
	limit, err := newsQueryInt(r, "limit", 50)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer")
		return
	}
	offset, err := newsQueryInt(r, "offset", 0)
	if err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "offset must be an integer")
		return
	}
	items, err := h.service.ListAdmin(r.Context(), r.URL.Query().Get("status"), limit, offset)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), r.PathValue("slug"))
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input Input
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	principal, _ := middleware.PrincipalFromContext(r.Context())
	item, err := h.service.Create(r.Context(), principal.UserID, input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	var input Input
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	item, err := h.service.Update(r.Context(), r.PathValue("postID"), input)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) SetStatus(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(w, r, &input); err != nil {
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
		return
	}
	item, err := h.service.SetStatus(r.Context(), r.PathValue("postID"), input.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("postID")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func newsQueryInt(r *http.Request, key string, fallback int) (int, error) {
	if !r.URL.Query().Has(key) {
		return fallback, nil
	}
	return strconv.Atoi(r.URL.Query().Get(key))
}

func decodeJSON(w http.ResponseWriter, r *http.Request, value any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 256<<10))
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
		slog.Error("write news response", "error", err)
	}
}

func writeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalid):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "invalid news post")
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "news post not found")
	case errors.Is(err, ErrConflict):
		httpapi.WriteError(w, http.StatusConflict, "slug_conflict", "a news post already uses this slug")
	default:
		slog.Error("news operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
