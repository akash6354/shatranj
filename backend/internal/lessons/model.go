package lessons

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotFound     = errors.New("lesson not found")
	ErrInvalidInput = errors.New("invalid lesson request")
)

type Chapter struct {
	ID              string          `json:"id"`
	CourseID        string          `json:"course_id"`
	Title           string          `json:"title"`
	Description     string          `json:"description,omitempty"`
	SortOrder       int             `json:"sort_order"`
	Content         json.RawMessage `json:"content"`
	ProgressPercent int             `json:"progress_percent"`
	Completed       bool            `json:"completed"`
	LastPosition    json.RawMessage `json:"last_position,omitempty"`
	UpdatedAt       *time.Time      `json:"updated_at,omitempty"`
}

type Course struct {
	ID              string    `json:"id"`
	Title           string    `json:"title"`
	Description     string    `json:"description,omitempty"`
	Category        string    `json:"category"`
	CoverImageURL   string    `json:"cover_image_url,omitempty"`
	SortOrder       int       `json:"sort_order"`
	ProgressPercent int       `json:"progress_percent"`
	Chapters        []Chapter `json:"chapters"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ProgressInput struct {
	ProgressPercent int             `json:"progress_percent"`
	Completed       bool            `json:"completed"`
	LastPosition    json.RawMessage `json:"last_position,omitempty"`
}

type Repository interface {
	List(context.Context, string) ([]Course, error)
	Get(context.Context, string, string) (Course, error)
	UpdateProgress(context.Context, string, string, ProgressInput) (Chapter, error)
}
