package tournaments

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type Handler struct{ service *Service }

func NewHandler(service *Service) *Handler { return &Handler{service: service} }

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.service.List(r.Context())
	if err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusOK, rows)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := decodeTournamentJSON(w, r, &input); err != nil {
		writeTournamentDecodeError(w, err)
		return
	}
	tournament, err := h.service.Create(r.Context(), tournamentUserID(r), input)
	if err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusCreated, tournament)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	tournament, err := h.service.Get(r.Context(), r.PathValue("tournamentID"))
	if err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusOK, tournament)
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Register(r.Context(), r.PathValue("tournamentID"), tournamentUserID(r)); err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusOK, map[string]string{"status": "registered"})
}

func (h *Handler) Withdraw(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Withdraw(r.Context(), r.PathValue("tournamentID"), tournamentUserID(r)); err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusOK, map[string]string{"status": "withdrawn"})
}

func (h *Handler) StartRound(w http.ResponseWriter, r *http.Request) {
	round, err := h.service.StartRound(r.Context(), r.PathValue("tournamentID"), tournamentUserID(r))
	if err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusCreated, round)
}

func (h *Handler) Standings(w http.ResponseWriter, r *http.Request) {
	standings, err := h.service.Standings(r.Context(), r.PathValue("tournamentID"))
	if err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusOK, standings)
}

func (h *Handler) ReportResult(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Result string `json:"result"`
	}
	if err := decodeTournamentJSON(w, r, &input); err != nil {
		writeTournamentDecodeError(w, err)
		return
	}
	if err := h.service.ReportResult(r.Context(), r.PathValue("tournamentID"),
		r.PathValue("pairingID"), tournamentUserID(r), input.Result); err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusOK, map[string]string{"status": "result_recorded"})
}

func (h *Handler) Complete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Complete(r.Context(), r.PathValue("tournamentID"), tournamentUserID(r)); err != nil {
		writeTournamentError(w, err)
		return
	}
	writeTournamentJSON(w, http.StatusOK, map[string]string{"status": "completed"})
}

func tournamentUserID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}

func decodeTournamentJSON(w http.ResponseWriter, r *http.Request, value any) error {
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

func writeTournamentJSON(w http.ResponseWriter, status int, value any) {
	if err := httpapi.WriteJSON(w, status, value); err != nil {
		slog.Error("write tournament response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeTournamentDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
}

func writeTournamentError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "tournament_not_found", "tournament not found")
	case errors.Is(err, ErrForbidden):
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "only the tournament creator can start rounds")
	case errors.Is(err, ErrInvalidRequest):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	case errors.Is(err, ErrConflict):
		httpapi.WriteError(w, http.StatusConflict, "tournament_conflict", "tournament is not in a state for this action")
	default:
		slog.Error("tournament operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
