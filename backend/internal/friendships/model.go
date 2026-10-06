package friendships

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("friendship not found")
	ErrConflict = errors.New("friendship state conflict")
	ErrInvalid  = errors.New("invalid friendship request")
)

type Friendship struct {
	UserID      string    `json:"user_id"`
	Username    string    `json:"username,omitempty"`
	Status      string    `json:"status"`
	RequestedBy string    `json:"requested_by"`
	CreatedAt   time.Time `json:"created_at"`
}

type Repository interface {
	List(context.Context, string) ([]Friendship, error)
	Requests(context.Context, string) ([]Friendship, error)
	Request(context.Context, string, string) error
	Accept(context.Context, string, string) error
	Reject(context.Context, string, string) error
	Block(context.Context, string, string) error
	Unblock(context.Context, string, string) error
	DirectAllowed(context.Context, string, string) error
}
