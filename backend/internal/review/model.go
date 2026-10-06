package review

import (
	"context"
	"errors"
	"time"

	"github.com/shatranj/backend/internal/analysis"
	"github.com/shatranj/backend/internal/games"
)

var (
	ErrNotFound        = errors.New("review not found")
	ErrForbidden       = errors.New("user is not a player in this game")
	ErrGameNotFinished = errors.New("game is not finished")
)

type Status string

const (
	StatusPending     Status = "pending"
	StatusRunning     Status = "running"
	StatusCompleted   Status = "completed"
	StatusUnavailable Status = "unavailable"
	StatusFailed      Status = "failed"
)

type Review struct {
	ID        string           `json:"id"`
	GameID    string           `json:"game_id"`
	Status    Status           `json:"status"`
	Result    *analysis.Result `json:"result,omitempty"`
	Message   string           `json:"message,omitempty"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
}

type Repository interface {
	Request(context.Context, string, string, bool) (Review, error)
	Get(context.Context, string) (Review, error)
}

type GameReader interface {
	Get(context.Context, string) (games.Game, error)
}

type JobRepository interface {
	analysis.Store
}
