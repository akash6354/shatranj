package chat

import (
	"context"
	"errors"
	"strings"
)

type Service struct {
	repository Repository
	friends    FriendValidator
	publisher  Publisher
	moderation ModerationHook
}

func NewService(repository Repository, friends FriendValidator, publisher Publisher, moderation ModerationHook) *Service {
	if moderation == nil {
		moderation = BasicModeration{}
	}
	return &Service{repository: repository, friends: friends, publisher: publisher, moderation: moderation}
}

func (s *Service) ListRooms(ctx context.Context, userID string) ([]Room, error) {
	if !validID(userID) {
		return nil, ErrForbidden
	}
	return s.repository.ListRooms(ctx, userID)
}

func (s *Service) Open(ctx context.Context, userID string, input CreateRoomInput) (Room, error) {
	if !validID(userID) {
		return Room{}, ErrForbidden
	}
	if input.Kind == DirectRoom {
		if !validID(input.UserID) || input.UserID == userID {
			return Room{}, ErrInvalid
		}
		if s.friends == nil {
			return Room{}, ErrForbidden
		}
		if err := s.friends.DirectAllowed(ctx, userID, input.UserID); err != nil {
			if errors.Is(err, ErrInvalid) {
				return Room{}, ErrInvalid
			}
			return Room{}, ErrForbidden
		}
	} else if (input.Kind == GameRoom || input.Kind == ClubRoom) && !validID(input.ReferenceID) {
		return Room{}, ErrInvalid
	}
	return s.repository.Open(ctx, userID, input)
}

func (s *Service) Authorize(ctx context.Context, roomID, userID string) error {
	if !validID(roomID) || !validID(userID) {
		return ErrForbidden
	}
	return s.repository.Authorize(ctx, roomID, userID)
}

func (s *Service) Messages(ctx context.Context, roomID, userID string, limit int) ([]Message, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if err := s.Authorize(ctx, roomID, userID); err != nil {
		return nil, err
	}
	return s.repository.Messages(ctx, roomID, userID, limit)
}

func (s *Service) Send(ctx context.Context, roomID, userID, content string) (Message, error) {
	content = strings.TrimSpace(content)
	if err := s.Authorize(ctx, roomID, userID); err != nil {
		return Message{}, err
	}
	if err := s.moderation.Check(ctx, roomID, userID, content); err != nil {
		return Message{}, err
	}
	message, err := s.repository.Send(ctx, roomID, userID, content)
	if err != nil {
		return Message{}, err
	}
	if s.publisher != nil {
		s.publisher.PublishRoom(roomID, map[string]any{"type": "chat.message", "message": message})
	}
	return message, nil
}

func (s *Service) Hide(ctx context.Context, roomID, messageID, moderatorID, reason string) error {
	if !validID(messageID) {
		return ErrInvalid
	}
	if err := s.Authorize(ctx, roomID, moderatorID); err != nil {
		return err
	}
	reason = strings.TrimSpace(reason)
	if reason == "" || len(reason) > 300 {
		return ErrInvalid
	}
	return s.repository.Hide(ctx, roomID, messageID, moderatorID, reason)
}

func (s *Service) SendChat(ctx context.Context, roomID, userID, content string) (any, error) {
	return s.Send(ctx, roomID, userID, content)
}

func (s *Service) AuthorizeChat(ctx context.Context, roomID, userID string) error {
	return s.Authorize(ctx, roomID, userID)
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
