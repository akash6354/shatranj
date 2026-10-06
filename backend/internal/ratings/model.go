package ratings

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidResult   = errors.New("invalid game result")
	ErrUnsupportedMode = errors.New("unsupported rating mode")
)

type Mode string

const (
	ModeBullet Mode = "bullet"
	ModeBlitz  Mode = "blitz"
	ModeRapid  Mode = "rapid"
	ModePuzzle Mode = "puzzle"
)

func (m Mode) Valid() bool {
	return m == ModeBullet || m == ModeBlitz || m == ModeRapid || m == ModePuzzle
}

type Record struct {
	UserID    string    `json:"user_id"`
	Mode      Mode      `json:"mode"`
	Rating    int       `json:"rating"`
	Highest   int       `json:"highest"`
	Games     int       `json:"games"`
	Wins      int       `json:"wins"`
	Draws     int       `json:"draws"`
	Losses    int       `json:"losses"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GameResult struct {
	GameID      string
	WhiteUserID string
	BlackUserID string
	Mode        Mode
	Result      string
}

type Change struct {
	UserID      string `json:"user_id"`
	Mode        Mode   `json:"mode"`
	OldRating   int    `json:"old_rating"`
	NewRating   int    `json:"new_rating"`
	RatingDelta int    `json:"rating_delta"`
}

type Repository interface {
	Get(context.Context, string, Mode) (Record, error)
	RecordGame(context.Context, GameResult) error
}
