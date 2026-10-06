package friendships

import (
	"context"
	"strings"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, userID string) ([]Friendship, error) {
	if userID == "" {
		return nil, ErrInvalid
	}
	return s.repository.List(ctx, userID)
}
func (s *Service) Requests(ctx context.Context, userID string) ([]Friendship, error) {
	if userID == "" {
		return nil, ErrInvalid
	}
	return s.repository.Requests(ctx, userID)
}
func (s *Service) Request(ctx context.Context, userID, targetID string) error {
	if !validID(userID) || !validID(targetID) || userID == targetID {
		return ErrInvalid
	}
	return s.repository.Request(ctx, userID, targetID)
}
func (s *Service) Accept(ctx context.Context, userID, requesterID string) error {
	if !validID(userID) || !validID(requesterID) || userID == requesterID {
		return ErrInvalid
	}
	return s.repository.Accept(ctx, userID, requesterID)
}
func (s *Service) Reject(ctx context.Context, userID, requesterID string) error {
	if !validID(userID) || !validID(requesterID) || userID == requesterID {
		return ErrInvalid
	}
	return s.repository.Reject(ctx, userID, requesterID)
}
func (s *Service) Block(ctx context.Context, userID, targetID string) error {
	if !validID(userID) || !validID(targetID) || userID == targetID {
		return ErrInvalid
	}
	return s.repository.Block(ctx, userID, targetID)
}
func (s *Service) Unblock(ctx context.Context, userID, targetID string) error {
	if !validID(userID) || !validID(targetID) || userID == targetID {
		return ErrInvalid
	}
	return s.repository.Unblock(ctx, userID, targetID)
}

func (s *Service) DirectAllowed(ctx context.Context, userID, targetID string) error {
	if !validID(userID) || !validID(targetID) || userID == targetID {
		return ErrInvalid
	}
	return s.repository.DirectAllowed(ctx, userID, targetID)
}

func validID(id string) bool {
	id = strings.TrimSpace(id)
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
