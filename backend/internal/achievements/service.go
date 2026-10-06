package achievements

import (
	"context"
	"strings"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Catalog(ctx context.Context) ([]Definition, error) {
	return s.repository.Catalog(ctx)
}

func (s *Service) List(ctx context.Context, userID string) ([]Achievement, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, ErrInvalid
	}
	return s.repository.List(ctx, userID)
}

// HandleEvent accumulates event progress and returns newly unlocked awards.
func (s *Service) HandleEvent(ctx context.Context, event Event) ([]Achievement, error) {
	event.UserID = strings.TrimSpace(event.UserID)
	event.Type = strings.TrimSpace(event.Type)
	event.SourceID = strings.TrimSpace(event.SourceID)
	if event.UserID == "" || event.Type == "" || len(event.Type) > 80 ||
		event.SourceID == "" || len(event.SourceID) > 160 || event.Amount < 1 || event.Amount > 1_000_000 {
		return nil, ErrInvalid
	}
	return s.repository.ProcessEvent(ctx, event)
}
