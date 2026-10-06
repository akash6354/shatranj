package chat

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound  = errors.New("chat room or message not found")
	ErrForbidden = errors.New("chat access forbidden")
	ErrInvalid   = errors.New("invalid chat request")
	ErrModerated = errors.New("message rejected by moderation")
)

type Kind string

const (
	GameRoom   Kind = "game"
	ClubRoom   Kind = "club"
	DirectRoom Kind = "direct"
)

type Room struct {
	ID        string    `json:"id"`
	Kind      Kind      `json:"kind"`
	Reference string    `json:"reference_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Message struct {
	ID        string    `json:"id"`
	RoomID    string    `json:"room_id"`
	SenderID  string    `json:"sender_id"`
	Username  string    `json:"username,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}

type SendInput struct {
	Content string `json:"content"`
}

type CreateRoomInput struct {
	Kind        Kind   `json:"kind"`
	ReferenceID string `json:"reference_id,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

type Repository interface {
	ListRooms(context.Context, string) ([]Room, error)
	Open(context.Context, string, CreateRoomInput) (Room, error)
	Authorize(context.Context, string, string) error
	Messages(context.Context, string, string, int) ([]Message, error)
	Send(context.Context, string, string, string) (Message, error)
	Hide(context.Context, string, string, string, string) error
}

type FriendValidator interface {
	DirectAllowed(context.Context, string, string) error
}

type Publisher interface {
	PublishRoom(string, any)
}

type ModerationHook interface {
	Check(context.Context, string, string, string) error
}
