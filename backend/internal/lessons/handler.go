package lessons

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/akash6354/shatranj/backend/internal/httpapi"
	"github.com/akash6354/shatranj/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	courses, err := h.service.List(r.Context(), lessonPrincipalID(r))
	if err != nil {
		writeLessonError(w, err)
		return
	}
	writeLessonJSON(w, http.StatusOK, courses)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	course, err := h.service.Get(r.Context(), lessonPrincipalID(r), r.PathValue("courseID"))
	if err != nil {
		writeLessonError(w, err)
		return
	}
	writeLessonJSON(w, http.StatusOK, course)
}

func (h *Handler) UpdateProgress(w http.ResponseWriter, r *http.Request) {
	var input ProgressInput
	if err := decodeLessonJSON(w, r, &input); err != nil {
		writeLessonDecodeError(w, err)
		return
	}
	chapter, err := h.service.UpdateProgress(
		r.Context(), lessonPrincipalID(r), r.PathValue("chapterID"), input,
	)
	if err != nil {
		writeLessonError(w, err)
		return
	}
	writeLessonJSON(w, http.StatusOK, chapter)
}

func lessonPrincipalID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}

func decodeLessonJSON(w http.ResponseWriter, r *http.Request, value any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(value); err != nil {
		return err
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return errors.New("request must contain one JSON object")
	}
	return nil
}

func writeLessonJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write lesson response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeLessonDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
}

func writeLessonError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "lesson_not_found", "lesson not found")
	case errors.Is(err, ErrInvalidInput):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	default:
		slog.Error("lesson operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
