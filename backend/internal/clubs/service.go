package clubs

import (
	"context"
	"fmt"
	"regexp"
	"strings"
)

var clubNamePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9 _-]{2,79}$`)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, userID string) ([]Club, error) {
	return s.repository.List(ctx, userID)
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Club, error) {
	input.Name = strings.TrimSpace(input.Name)
	input.Description = strings.TrimSpace(input.Description)
	if userID == "" || !clubNamePattern.MatchString(input.Name) || len(input.Description) > 1000 {
		return Club{}, fmt.Errorf("%w: invalid name or description", ErrInvalid)
	}
	if input.Visibility == "" {
		input.Visibility = "public"
	}
	if input.Visibility != "public" && input.Visibility != "private" {
		return Club{}, ErrInvalid
	}
	return s.repository.Create(ctx, userID, input)
}

func (s *Service) Get(ctx context.Context, clubID, userID string) (Club, error) {
	if !validClubID(clubID) {
		return Club{}, ErrNotFound
	}
	return s.repository.Get(ctx, clubID, userID)
}

func (s *Service) Members(ctx context.Context, clubID, userID string) ([]Member, error) {
	if !validClubID(clubID) {
		return nil, ErrNotFound
	}
	if err := s.repository.Authorize(ctx, clubID, userID); err != nil {
		return nil, err
	}
	return s.repository.Members(ctx, clubID, userID)
}

func (s *Service) Join(ctx context.Context, clubID, userID string) (string, error) {
	if !validClubID(clubID) || userID == "" {
		return "", ErrInvalid
	}
	return s.repository.Join(ctx, clubID, userID)
}

func (s *Service) Leave(ctx context.Context, clubID, userID string) error {
	if !validClubID(clubID) || userID == "" {
		return ErrInvalid
	}
	return s.repository.Leave(ctx, clubID, userID)
}

func (s *Service) Requests(ctx context.Context, clubID, userID string) ([]JoinRequest, error) {
	if !validClubID(clubID) || userID == "" {
		return nil, ErrInvalid
	}
	return s.repository.Requests(ctx, clubID, userID)
}

func (s *Service) ResolveRequest(ctx context.Context, clubID, requestID, adminID string, approve bool) error {
	if !validClubID(clubID) || !validClubID(requestID) || adminID == "" {
		return ErrInvalid
	}
	return s.repository.ResolveRequest(ctx, clubID, requestID, adminID, approve)
}

func (s *Service) SetRole(ctx context.Context, clubID, targetID, adminID, role string) error {
	if !validClubID(clubID) || !validClubID(targetID) ||
		adminID == "" ||
		(role != "admin" && role != "moderator" && role != "member") {
		return ErrInvalid
	}
	return s.repository.SetRole(ctx, clubID, targetID, adminID, role)
}

func validClubID(id string) bool {
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
