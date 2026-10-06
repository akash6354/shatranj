package puzzles

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound      = errors.New("puzzle not found")
	ErrInvalidInput  = errors.New("invalid puzzle request")
	ErrInvalidPuzzle = errors.New("stored puzzle is invalid")
	ErrRushNotFound  = errors.New("puzzle rush session not found")
	ErrRushFinished  = errors.New("puzzle rush session is not active")
	ErrRushPuzzle    = errors.New("puzzle does not belong to this rush session")
)

type Status string

const (
	StatusDraft     Status = "draft"
	StatusPublished Status = "published"
	StatusArchived  Status = "archived"
)

type Puzzle struct {
	ID          string    `json:"id"`
	FEN         string    `json:"fen"`
	Solution    []string  `json:"-"`
	Themes      []string  `json:"themes"`
	Difficulty  string    `json:"difficulty"`
	Rating      int       `json:"rating"`
	Explanation string    `json:"explanation,omitempty"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Filter struct {
	Theme      string
	Difficulty string
	Limit      int
	Offset     int
}

type AttemptInput struct {
	UserID        string
	PuzzleID      string
	RushSessionID string
	Moves         []string
	Correct       bool
}

type Attempt struct {
	ID                 string    `json:"id"`
	PuzzleID           string    `json:"puzzle_id"`
	Correct            bool      `json:"correct"`
	Moves              []string  `json:"moves"`
	RatingBefore       int       `json:"rating_before"`
	RatingAfter        int       `json:"rating_after"`
	PuzzleRatingBefore int       `json:"puzzle_rating_before"`
	PuzzleRatingAfter  int       `json:"puzzle_rating_after"`
	CreatedAt          time.Time `json:"created_at"`
}

type RushStatus string

const (
	RushActive   RushStatus = "active"
	RushFinished RushStatus = "finished"
	RushExpired  RushStatus = "expired"
)

type RushSession struct {
	ID               string     `json:"id"`
	UserID           string     `json:"-"`
	Status           RushStatus `json:"status"`
	Score            int        `json:"score"`
	Streak           int        `json:"streak"`
	TimeLimitSeconds int        `json:"time_limit_seconds"`
	CurrentPuzzleID  string     `json:"current_puzzle_id,omitempty"`
	CurrentPuzzle    *Puzzle    `json:"current_puzzle,omitempty"`
	StartedAt        time.Time  `json:"started_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	FinishedAt       *time.Time `json:"finished_at,omitempty"`
}

type RushAnswer struct {
	PuzzleID string   `json:"puzzle_id"`
	Moves    []string `json:"moves"`
}

type Repository interface {
	List(context.Context, Filter) ([]Puzzle, error)
	Get(context.Context, string) (Puzzle, error)
	Daily(context.Context, time.Time) (Puzzle, error)
	Rating(context.Context, string) (int, error)
	RecordAttempt(context.Context, AttemptInput) (Attempt, error)
	StartRush(context.Context, string, int) (RushSession, error)
	GetRush(context.Context, string, string) (RushSession, error)
	SubmitRushAnswer(context.Context, string, string, string, []string, bool) (RushSession, Attempt, error)
}
