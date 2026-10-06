package tournaments

import (
	"context"
	"fmt"
	"strings"
	"time"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context) ([]Tournament, error) {
	return s.repository.List(ctx)
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Tournament, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	input.Format = normalizeFormat(strings.ToLower(strings.TrimSpace(input.Format)))
	if userID == "" || input.Name == "" || len(input.Name) > 120 ||
		(input.Format != "swiss" && input.Format != "arena") ||
		input.TimeControlSecs < 30 || input.TimeControlSecs > 86400 ||
		input.IncrementSeconds < 0 || input.IncrementSeconds > 3600 ||
		input.MaxPlayers < 2 || input.MaxPlayers > 512 ||
		input.StartsAt.Before(time.Now().Add(-time.Minute)) {
		return Tournament{}, fmt.Errorf("%w: invalid tournament details", ErrInvalidRequest)
	}
	return s.repository.Create(ctx, userID, input)
}

func (s *Service) Get(ctx context.Context, id string) (Tournament, error) {
	if !validID(id) {
		return Tournament{}, ErrNotFound
	}
	return s.repository.Get(ctx, id)
}

func (s *Service) Register(ctx context.Context, id, userID string) error {
	if !validID(id) || userID == "" {
		return ErrInvalidRequest
	}
	return s.repository.Register(ctx, id, userID)
}

func (s *Service) Withdraw(ctx context.Context, id, userID string) error {
	if !validID(id) || userID == "" {
		return ErrInvalidRequest
	}
	return s.repository.Withdraw(ctx, id, userID)
}

func (s *Service) StartRound(ctx context.Context, id, userID string) (Round, error) {
	if !validID(id) || userID == "" {
		return Round{}, ErrInvalidRequest
	}
	return s.repository.StartRound(ctx, id, userID)
}

func (s *Service) Standings(ctx context.Context, id string) ([]Player, error) {
	if !validID(id) {
		return nil, ErrNotFound
	}
	return s.repository.Standings(ctx, id)
}

func (s *Service) ReportResult(ctx context.Context, tournamentID, pairingID, userID, result string) error {
	if !validID(tournamentID) || !validID(pairingID) || userID == "" ||
		(result != "1-0" && result != "0-1" && result != "1/2-1/2") {
		return ErrInvalidRequest
	}
	return s.repository.ReportResult(ctx, tournamentID, pairingID, userID, result)
}

func (s *Service) Complete(ctx context.Context, tournamentID, userID string) error {
	if !validID(tournamentID) || userID == "" {
		return ErrInvalidRequest
	}
	return s.repository.Complete(ctx, tournamentID, userID)
}

func validID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for index, value := range id {
		switch index {
		case 8, 13, 18, 23:
			if value != '-' {
				return false
			}
		default:
			if !(value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F') {
				return false
			}
		}
	}
	return true
}
