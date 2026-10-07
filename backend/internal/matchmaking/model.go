package matchmaking

import (
	"context"
	"errors"
	"time"

	"github.com/akash6354/shatranj/backend/internal/games"
	"github.com/akash6354/shatranj/backend/internal/ratings"
)

var (
	ErrAlreadyQueued   = errors.New("user already has an active matchmaking entry")
	ErrQueueNotFound   = errors.New("matchmaking entry not found")
	ErrInvalidQueue    = errors.New("invalid matchmaking request")
	ErrMatchInProgress = errors.New("match is already in progress")
)

type Status string

const (
	StatusQueued   Status = "queued"
	StatusMatched  Status = "matched"
	StatusPaired   Status = "paired"
	StatusCanceled Status = "cancelled"
)

type Entry struct {
	ID          string            `json:"id"`
	UserID      string            `json:"-"`
	Mode        ratings.Mode      `json:"mode"`
	TimeControl games.TimeControl `json:"time_control"`
	Rating      int               `json:"rating"`
	RatingRange int               `json:"rating_range"`
	Status      Status            `json:"status"`
	MatchID     string            `json:"match_id,omitempty"`
	GameID      string            `json:"game_id,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
}

type Request struct {
	Mode        ratings.Mode      `json:"mode"`
	TimeControl games.TimeControl `json:"time_control"`
}

type QueueRepository interface {
	JoinQueue(context.Context, Entry) (Entry, *Entry, error)
	LeaveQueue(context.Context, string) error
	GetQueueStatus(context.Context, string) (Entry, error)
	CompleteMatch(context.Context, string, string) error
	AbortMatch(context.Context, string) error
}

type RatingReader interface {
	Get(context.Context, string, ratings.Mode) (ratings.Record, error)
}

type GameCreator interface {
	Create(context.Context, string, games.CreateInput) (games.Game, error)
	Join(context.Context, string, string) (games.Game, error)
}
