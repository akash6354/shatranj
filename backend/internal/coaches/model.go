package coaches

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid   = errors.New("invalid coach request")
	ErrNotFound  = errors.New("coach or booking not found")
	ErrConflict  = errors.New("coach booking conflict")
	ErrForbidden = errors.New("coach action forbidden")
)

type Coach struct {
	UserID       string    `json:"user_id"`
	Username     string    `json:"username"`
	DisplayName  string    `json:"display_name"`
	Title        string    `json:"title"`
	Bio          string    `json:"bio"`
	RatePaise    int64     `json:"rate_paise"`
	Currency     string    `json:"currency"`
	Specialties  []string  `json:"specialties"`
	Availability any       `json:"availability"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type ProfileInput struct {
	Title        string   `json:"title"`
	Bio          string   `json:"bio"`
	RatePaise    int64    `json:"rate_paise"`
	Specialties  []string `json:"specialties"`
	Availability any      `json:"availability"`
}

type Booking struct {
	ID          string    `json:"id"`
	CoachID     string    `json:"coach_id"`
	StudentID   string    `json:"student_id"`
	StartsAt    time.Time `json:"starts_at"`
	DurationMin int       `json:"duration_minutes"`
	Notes       string    `json:"notes,omitempty"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BookingInput struct {
	StartsAt    time.Time `json:"starts_at"`
	DurationMin int       `json:"duration_minutes"`
	Notes       string    `json:"notes"`
}

type Repository interface {
	List(context.Context, int, int) ([]Coach, error)
	Get(context.Context, string) (Coach, error)
	Upsert(context.Context, string, ProfileInput) (Coach, error)
	CreateBooking(context.Context, string, string, BookingInput) (Booking, error)
	ListBookings(context.Context, string) ([]Booking, error)
	UpdateBookingStatus(context.Context, string, string, string) (Booking, error)
	SetStatus(context.Context, string, string) error
}
