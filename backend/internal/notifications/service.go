package notifications

import (
	"context"
	"encoding/json"
	"strings"
)

type Service struct {
	repository Repository
	push       PushSender
}

func NewService(repository Repository, push PushSender) *Service {
	return &Service{repository: repository, push: push}
}

func (s *Service) Create(ctx context.Context, input CreateInput) (Notification, error) {
	input.Type = strings.TrimSpace(input.Type)
	input.Title = strings.TrimSpace(input.Title)
	input.Body = strings.TrimSpace(input.Body)
	if !validID(input.UserID) || input.Type == "" || len(input.Type) > 80 ||
		input.Title == "" || len(input.Title) > 160 || len(input.Body) > 2000 {
		return Notification{}, ErrInvalid
	}
	if input.Data != nil {
		raw, err := json.Marshal(input.Data)
		if err != nil || len(raw) > 8192 {
			return Notification{}, ErrInvalid
		}
		input.Data = json.RawMessage(raw)
	}
	notification, err := s.repository.Create(ctx, input)
	if err != nil {
		return Notification{}, err
	}
	if s.push != nil {
		if err := s.push.Send(ctx, notification); err != nil {
			return notification, err
		}
	}
	return notification, nil
}

func (s *Service) List(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]Notification, error) {
	if !validID(userID) {
		return nil, ErrInvalid
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		return nil, ErrInvalid
	}
	return s.repository.List(ctx, userID, unreadOnly, limit, offset)
}

func (s *Service) UnreadCount(ctx context.Context, userID string) (int, error) {
	if !validID(userID) {
		return 0, ErrInvalid
	}
	return s.repository.UnreadCount(ctx, userID)
}

func (s *Service) SetRead(ctx context.Context, userID, notificationID string, read bool) (Notification, error) {
	if !validID(userID) || !validID(notificationID) {
		return Notification{}, ErrInvalid
	}
	return s.repository.SetRead(ctx, userID, notificationID, read)
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
