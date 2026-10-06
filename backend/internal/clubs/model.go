package clubs

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("club not found")
	ErrForbidden = errors.New("club action forbidden")
	ErrConflict  = errors.New("club membership conflict")
	ErrInvalid   = errors.New("invalid club request")
)

type Club struct {
	ID          string    `json:"id"`
	OwnerID     string    `json:"owner_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Visibility  string    `json:"visibility"`
	Role        string    `json:"role,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

type Member struct {
	UserID   string    `json:"user_id"`
	Username string    `json:"username"`
	Role     string    `json:"role"`
	JoinedAt time.Time `json:"joined_at"`
}

type JoinRequest struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Username  string    `json:"username"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateInput struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Visibility  string `json:"visibility"`
}

type Repository interface {
	List(context.Context, string) ([]Club, error)
	Create(context.Context, string, CreateInput) (Club, error)
	Get(context.Context, string, string) (Club, error)
	Members(context.Context, string, string) ([]Member, error)
	Join(context.Context, string, string) (string, error)
	Leave(context.Context, string, string) error
	Requests(context.Context, string, string) ([]JoinRequest, error)
	ResolveRequest(context.Context, string, string, string, bool) error
	SetRole(context.Context, string, string, string, string) error
	Authorize(context.Context, string, string) error
}
