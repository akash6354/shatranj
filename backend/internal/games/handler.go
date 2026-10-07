package games

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/shatranj/backend/internal/httpapi"
	"github.com/shatranj/backend/internal/middleware"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var input CreateInput
	if err := decodeGameJSON(w, r, &input); err != nil {
		writeGameDecodeError(w, err)
		return
	}
	game, err := h.service.Create(r.Context(), principalID(r), input)
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusCreated, game)
}

func (h *Handler) Join(w http.ResponseWriter, r *http.Request) {
	game, err := h.service.Join(r.Context(), r.PathValue("gameID"), principalID(r))
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusOK, game)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	game, err := h.service.Get(r.Context(), r.PathValue("gameID"), principalID(r))
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusOK, game)
}

func (h *Handler) Move(w http.ResponseWriter, r *http.Request) {
	var input MoveInput
	if err := decodeGameJSON(w, r, &input); err != nil {
		writeGameDecodeError(w, err)
		return
	}
	game, err := h.service.SubmitMove(r.Context(), r.PathValue("gameID"), principalID(r), input)
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusOK, game)
}

func (h *Handler) Resign(w http.ResponseWriter, r *http.Request) {
	game, err := h.service.Resign(r.Context(), r.PathValue("gameID"), principalID(r))
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusOK, game)
}

func (h *Handler) OfferDraw(w http.ResponseWriter, r *http.Request) {
	game, err := h.service.OfferDraw(r.Context(), r.PathValue("gameID"), principalID(r))
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusOK, game)
}

func (h *Handler) AcceptDraw(w http.ResponseWriter, r *http.Request) {
	game, err := h.service.AcceptDraw(r.Context(), r.PathValue("gameID"), principalID(r))
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusOK, game)
}

func (h *Handler) ClaimDraw(w http.ResponseWriter, r *http.Request) {
	var input DrawClaimInput
	if err := decodeGameJSON(w, r, &input); err != nil {
		writeGameDecodeError(w, err)
		return
	}
	game, err := h.service.ClaimDraw(r.Context(), r.PathValue("gameID"), principalID(r), input.Reason)
	if err != nil {
		writeGameError(w, err)
		return
	}
	writeGameJSON(w, http.StatusOK, game)
}

func principalID(r *http.Request) string {
	principal, ok := middleware.PrincipalFromContext(r.Context())
	if !ok {
		return ""
	}
	return principal.UserID
}

func decodeGameJSON(w http.ResponseWriter, r *http.Request, value any) error {
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

func writeGameJSON(w http.ResponseWriter, status int, game Game) {
	if err := httpapi.WriteJSON(w, status, game); err != nil {
		slog.Error("write game response", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}

func writeGameDecodeError(w http.ResponseWriter, err error) {
	var maxBytesError *http.MaxBytesError
	if errors.As(err, &maxBytesError) {
		httpapi.WriteError(w, http.StatusRequestEntityTooLarge, "request_too_large", "request body exceeds the 1 MiB limit")
		return
	}
	httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", "request body must be valid JSON")
}

func writeGameError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrNotFound):
		httpapi.WriteError(w, http.StatusNotFound, "game_not_found", "game not found")
	case errors.Is(err, ErrForbidden):
		httpapi.WriteError(w, http.StatusForbidden, "forbidden", "you are not a player in this game")
	case errors.Is(err, ErrNotYourTurn):
		httpapi.WriteError(w, http.StatusConflict, "not_your_turn", "it is not your turn")
	case errors.Is(err, ErrGameNotActive):
		httpapi.WriteError(w, http.StatusConflict, "game_not_active", "game is not active")
	case errors.Is(err, ErrGameFull):
		httpapi.WriteError(w, http.StatusConflict, "game_full", "game is not available to join")
	case errors.Is(err, ErrCannotJoinOwn):
		httpapi.WriteError(w, http.StatusConflict, "cannot_join_own_game", "game creator cannot join as opponent")
	case errors.Is(err, ErrMoveConflict):
		httpapi.WriteError(w, http.StatusConflict, "position_changed", "game position changed; reload and retry")
	case errors.Is(err, ErrTimeExpired):
		httpapi.WriteError(w, http.StatusConflict, "clock_expired", "the player's clock expired before the move was accepted")
	case errors.Is(err, ErrDrawUnavailable):
		httpapi.WriteError(w, http.StatusConflict, "draw_unavailable", "no draw offer is pending")
	case errors.Is(err, ErrInvalidRequest):
		httpapi.WriteError(w, http.StatusBadRequest, "invalid_request", err.Error())
	default:
		slog.Error("game operation failed", "error", err)
		httpapi.WriteError(w, http.StatusInternalServerError, "internal_error", "internal server error")
	}
}
