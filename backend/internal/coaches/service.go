package coaches

import (
	"context"
	"encoding/json"
	"strings"
	"time"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, limit, offset int) ([]Coach, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		return nil, ErrInvalid
	}
	return s.repository.List(ctx, limit, offset)
}

func (s *Service) Get(ctx context.Context, coachID string) (Coach, error) {
	if !validID(coachID) {
		return Coach{}, ErrNotFound
	}
	return s.repository.Get(ctx, coachID)
}

func (s *Service) Upsert(ctx context.Context, userID string, input ProfileInput) (Coach, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Bio = strings.TrimSpace(input.Bio)
	if !validID(userID) || input.Title == "" || len(input.Title) > 100 ||
		len(input.Bio) > 3000 || input.RatePaise < 0 || input.RatePaise > 100_000_000 ||
		len(input.Specialties) > 20 {
		return Coach{}, ErrInvalid
	}
	for i, item := range input.Specialties {
		input.Specialties[i] = strings.TrimSpace(item)
		if input.Specialties[i] == "" || len(input.Specialties[i]) > 60 {
			return Coach{}, ErrInvalid
		}
	}
	if input.Availability != nil {
		encoded, err := json.Marshal(input.Availability)
		if err != nil || len(encoded) > 16*1024 {
			return Coach{}, ErrInvalid
		}
	}
	return s.repository.Upsert(ctx, userID, input)
}

func (s *Service) Book(ctx context.Context, studentID, coachID string, input BookingInput) (Booking, error) {
	input.Notes = strings.TrimSpace(input.Notes)
	if !validID(studentID) || !validID(coachID) || studentID == coachID ||
		input.StartsAt.Before(time.Now().Add(-time.Minute)) ||
		input.DurationMin < 15 || input.DurationMin > 240 || len(input.Notes) > 1000 {
		return Booking{}, ErrInvalid
	}
	return s.repository.CreateBooking(ctx, studentID, coachID, input)
}

func (s *Service) Bookings(ctx context.Context, userID string) ([]Booking, error) {
	if !validID(userID) {
		return nil, ErrInvalid
	}
	return s.repository.ListBookings(ctx, userID)
}

func (s *Service) UpdateBooking(ctx context.Context, userID, bookingID, status string) (Booking, error) {
	if !validID(userID) || !validID(bookingID) ||
		(status != "accepted" && status != "declined" && status != "cancelled") {
		return Booking{}, ErrInvalid
	}
	return s.repository.UpdateBookingStatus(ctx, userID, bookingID, status)
}

func (s *Service) SetStatus(ctx context.Context, coachID, status string) error {
	if !validID(coachID) || (status != "active" && status != "rejected" && status != "pending") {
		return ErrInvalid
	}
	return s.repository.SetStatus(ctx, coachID, status)
}

func validID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for index, value := range id {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			if value != '-' {
				return false
			}
		} else if !(value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F') {
			return false
		}
	}
	return true
}
