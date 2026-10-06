package notifications

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid notification")
	ErrNotFound = errors.New("notification not found")
)

type Notification struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id,omitempty"`
	Type      string     `json:"type"`
	Title     string     `json:"title"`
	Body      string     `json:"body"`
	Data      any        `json:"data,omitempty"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type CreateInput struct {
	UserID string
	Type   string
	Title  string
	Body   string
	Data   any
}

type Repository interface {
	Create(context.Context, CreateInput) (Notification, error)
	List(context.Context, string, bool, int, int) ([]Notification, error)
	UnreadCount(context.Context, string) (int, error)
	SetRead(context.Context, string, string, bool) (Notification, error)
}

type PushSender interface {
	Send(context.Context, Notification) error
}
