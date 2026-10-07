package puzzles

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

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := Filter{Theme: query.Get("theme"), Difficulty: query.Get("difficulty"), Limit: 20}
	var err error
	if query.Has("limit") {
		filter.Limit, err = strconv.Atoi(query.Get("limit"))
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "limit must be an integer")
			return
		}
	}
	if query.Has("offset") {
		filter.Offset, err = strconv.Atoi(query.Get("offset"))
		if err != nil {
			httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "offset must be an integer")
			return
		}
	}
	puzzles, err := h.service.List(r.Context(), filter)
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusOK, puzzles)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	puzzle, err := h.service.Get(r.Context(), r.PathValue("puzzleID"))
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusOK, puzzle)
}

func (h *Handler) Daily(w http.ResponseWriter, r *http.Request) {
	puzzle, err := h.service.Daily(r.Context())
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusOK, puzzle)
}

func (h *Handler) Rating(w http.ResponseWriter, r *http.Request) {
	rating, err := h.service.Rating(r.Context(), puzzlePrincipalID(r))
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusOK, map[string]int{"rating": rating})
}

func (h *Handler) Submit(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Moves []string `json:"moves"`
	}
	if err := decodePuzzleJSON(w, r, &input); err != nil {
		writePuzzleDecodeError(w, err)
		return
	}
	attempt, err := h.service.Submit(r.Context(), puzzlePrincipalID(r), r.PathValue("puzzleID"), input.Moves)
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusOK, attempt)
}

func (h *Handler) StartRush(w http.ResponseWriter, r *http.Request) {
	var input struct {
		TimeLimitSeconds int `json:"time_limit_seconds"`
	}
	if r.Body != nil && r.ContentLength != 0 {
		if err := decodePuzzleJSON(w, r, &input); err != nil {
			writePuzzleDecodeError(w, err)
			return
		}
	}
	session, err := h.service.StartRush(r.Context(), puzzlePrincipalID(r), input.TimeLimitSeconds)
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusCreated, session)
}

func (h *Handler) GetRush(w http.ResponseWriter, r *http.Request) {
	session, err := h.service.GetRush(r.Context(), puzzlePrincipalID(r), r.PathValue("sessionID"))
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusOK, session)
}

func (h *Handler) SubmitRushAnswer(w http.ResponseWriter, r *http.Request) {
	var answer RushAnswer
	if err := decodePuzzleJSON(w, r, &answer); err != nil {
		writePuzzleDecodeError(w, err)
		return
	}
	session, attempt, err := h.service.SubmitRushAnswer(
		r.Context(), puzzlePrincipalID(r), r.PathValue("sessionID"), answer,
	)
	if err != nil {
		writePuzzleError(w, err)
		return
	}
	writePuzzleJSON(w, http.StatusOK, struct {
		Session RushSession `json:"session"`
		Attempt Attempt     `json:"attempt"`
	}{Session: session, Attempt: attempt})
}

func puzzlePrincipalID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}

func decodePuzzleJSON(w http.ResponseWriter, r *http.Request, value any) error {
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

func writePuzzleJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write puzzle response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writePuzzleDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
}

func writePuzzleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound), errors.Is(err, ErrRushNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "not_found", "puzzle or rush session not found")
	case errors.Is(err, ErrInvalidInput):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, ErrRushFinished):
		httpapi.WriteError(w, http.StatusConflict, "rush_finished", "puzzle rush session is not active")
	case errors.Is(err, ErrRushPuzzle):
		httpapi.WriteError(w, http.StatusConflict, "wrong_rush_puzzle", "answer does not match the current puzzle")
	default:
		slog.Error("puzzle operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
