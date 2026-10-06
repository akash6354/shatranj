package achievements

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid achievement event")
	ErrNotFound = errors.New("achievement not found")
)

type Definition struct {
	Key         string    `json:"key"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	EventType   string    `json:"-"`
	Threshold   int       `json:"threshold"`
	CreatedAt   time.Time `json:"created_at"`
}

type Achievement struct {
	Definition
	UnlockedAt *time.Time `json:"unlocked_at,omitempty"`
	Progress   int        `json:"progress"`
	Unlocked   bool       `json:"unlocked"`
}

type Event struct {
	UserID   string
	Type     string
	SourceID string
	Amount   int
}

type Repository interface {
	Catalog(context.Context) ([]Definition, error)
	List(context.Context, string) ([]Achievement, error)
	ProcessEvent(context.Context, Event) ([]Achievement, error)
}
