package tournaments

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound       = errors.New("tournament not found")
	ErrForbidden      = errors.New("tournament action forbidden")
	ErrInvalidRequest = errors.New("invalid tournament request")
	ErrConflict       = errors.New("tournament state conflict")
)

type Status string

const (
	StatusUpcoming  Status = "upcoming"
	StatusLive      Status = "live"
	StatusCompleted Status = "completed"
	StatusCancelled Status = "cancelled"
)

type Tournament struct {
	ID               string    `json:"id"`
	CreatorID        string    `json:"-"`
	Name             string    `json:"name"`
	Description      string    `json:"description,omitempty"`
	Format           string    `json:"format"`
	Status           Status    `json:"status"`
	TimeControlSecs  int       `json:"time_control_seconds"`
	IncrementSeconds int       `json:"increment_seconds"`
	MaxPlayers       int       `json:"max_players"`
	StartsAt         time.Time `json:"starts_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Participants     []Player  `json:"participants,omitempty"`
	Rounds           []Round   `json:"rounds,omitempty"`
}

type Player struct {
	UserID   string `json:"user_id"`
	Username string `json:"username,omitempty"`
	Rating   int    `json:"rating"`
	Score    int    `json:"score"`
	Rank     int    `json:"rank,omitempty"`
}

type Round struct {
	Number   int       `json:"number"`
	Status   string    `json:"status"`
	Pairings []Pairing `json:"pairings"`
}

type Pairing struct {
	ID         string `json:"id"`
	WhiteID    string `json:"white_player_id,omitempty"`
	BlackID    string `json:"black_player_id,omitempty"`
	GameID     string `json:"game_id,omitempty"`
	Result     string `json:"result"`
	WhiteScore int    `json:"white_score"`
	BlackScore int    `json:"black_score"`
	Bye        bool   `json:"bye"`
}

type CreateInput struct {
	Name             string    `json:"name"`
	Description      string    `json:"description"`
	Format           string    `json:"format"`
	TimeControlSecs  int       `json:"time_control_seconds"`
	IncrementSeconds int       `json:"increment_seconds"`
	MaxPlayers       int       `json:"max_players"`
	StartsAt         time.Time `json:"starts_at"`
}

type Repository interface {
	List(context.Context) ([]Tournament, error)
	Create(context.Context, string, CreateInput) (Tournament, error)
	Get(context.Context, string) (Tournament, error)
	Register(context.Context, string, string) error
	Withdraw(context.Context, string, string) error
	StartRound(context.Context, string, string) (Round, error)
	Standings(context.Context, string) ([]Player, error)
	ReportResult(context.Context, string, string, string, string) error
	Complete(context.Context, string, string) error
}
