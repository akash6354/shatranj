package games

import (
	"errors"
	"time"

	"github.com/shatranj/backend/internal/chess"
)

var (
	ErrNotFound        = errors.New("game not found")
	ErrForbidden       = errors.New("user is not a player in this game")
	ErrNotYourTurn     = errors.New("it is not your turn")
	ErrGameNotActive   = errors.New("game is not active")
	ErrGameFull        = errors.New("game already has two players")
	ErrCannotJoinOwn   = errors.New("game creator cannot join as opponent")
	ErrMoveConflict    = errors.New("game position changed; reload and retry")
	ErrInvalidRequest  = errors.New("invalid game request")
	ErrDrawUnavailable = errors.New("no draw offer is pending")
)

type Status string

const (
	StatusWaiting  Status = "waiting"
	StatusActive   Status = "active"
	StatusFinished Status = "finished"
)

type Result string

const (
	ResultWhiteWin Result = "1-0"
	ResultBlackWin Result = "0-1"
	ResultDraw     Result = "1/2-1/2"
	ResultOngoing  Result = "*"
)

type Game struct {
	ID            string       `json:"id"`
	WhitePlayerID string       `json:"white_player_id"`
	BlackPlayerID string       `json:"black_player_id,omitempty"`
	TimeControl   TimeControl  `json:"time_control"`
	Mode          string       `json:"mode"`
	Rated         bool         `json:"rated"`
	CurrentFEN    string       `json:"current_fen"`
	Status        Status       `json:"status"`
	Result        Result       `json:"result"`
	EndReason     string       `json:"end_reason,omitempty"`
	DrawOfferBy   string       `json:"draw_offer_by,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	StartedAt     *time.Time   `json:"started_at,omitempty"`
	FinishedAt    *time.Time   `json:"finished_at,omitempty"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Moves         []MoveRecord `json:"moves,omitempty"`
}

type TimeControl struct {
	InitialSeconds   int `json:"initial_seconds"`
	IncrementSeconds int `json:"increment_seconds"`
}

type MoveRecord struct {
	Number       int       `json:"number"`
	PlayerID     string    `json:"player_id"`
	UCI          string    `json:"uci"`
	SAN          string    `json:"san"`
	FENAfter     string    `json:"fen_after"`
	WhiteClockMS *int64    `json:"white_clock_ms,omitempty"`
	BlackClockMS *int64    `json:"black_clock_ms,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateInput struct {
	TimeControl TimeControl `json:"time_control"`
	Mode        string      `json:"mode,omitempty"`
	Rated       bool        `json:"rated,omitempty"`
}

type MoveInput struct {
	UCI string `json:"uci"`
}

type GameEvent struct {
	Type      string      `json:"type"`
	GameID    string      `json:"game_id"`
	Game      *Game       `json:"game,omitempty"`
	Move      *MoveRecord `json:"move,omitempty"`
	PlayerID  string      `json:"player_id,omitempty"`
	Color     chess.Color `json:"color,omitempty"`
	Result    Result      `json:"result,omitempty"`
	Reason    string      `json:"reason,omitempty"`
	CreatedAt time.Time   `json:"created_at"`
}

type MoveUpdate struct {
	PlayerID string
	Move     chess.Move
	Record   MoveRecord
	FEN      string
	Status   Status
	Result   Result
	Reason   string
}
