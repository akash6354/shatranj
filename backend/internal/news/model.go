package news

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid  = errors.New("invalid news post")
	ErrNotFound = errors.New("news post not found")
	ErrConflict = errors.New("news slug already exists")
)

type Post struct {
	ID          string     `json:"id"`
	AuthorID    string     `json:"author_id,omitempty"`
	Title       string     `json:"title"`
	Slug        string     `json:"slug"`
	Excerpt     string     `json:"excerpt"`
	Body        string     `json:"body"`
	CoverURL    string     `json:"cover_url,omitempty"`
	Tags        []string   `json:"tags"`
	Status      string     `json:"status"`
	PublishedAt *time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type Input struct {
	Title    string   `json:"title"`
	Slug     string   `json:"slug"`
	Excerpt  string   `json:"excerpt"`
	Body     string   `json:"body"`
	CoverURL string   `json:"cover_url"`
	Tags     []string `json:"tags"`
}

type Repository interface {
	List(context.Context, int, int) ([]Post, error)
	ListAdmin(context.Context, string, int, int) ([]Post, error)
	Get(context.Context, string) (Post, error)
	Create(context.Context, string, Input) (Post, error)
	Update(context.Context, string, Input) (Post, error)
	SetStatus(context.Context, string, string) (Post, error)
	Delete(context.Context, string) error
}
