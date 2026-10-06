package chat

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestBasicModeration(t *testing.T) {
	moderation := BasicModeration{}
	tests := []struct {
		name    string
		content string
		want    error
	}{
		{name: "ordinary text", content: "Good game!"},
		{name: "empty", content: " \t ", want: ErrInvalid},
		{name: "too long", content: strings.Repeat("x", 2001), want: ErrInvalid},
		{name: "control character", content: "hello\x00world", want: ErrModerated},
		{name: "newline allowed", content: "hello\nworld"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := moderation.Check(context.Background(), "room", "user", test.content)
			if !errors.Is(err, test.want) {
				t.Fatalf("Check() error = %v, want %v", err, test.want)
			}
		})
	}
}

func TestServiceRejectsInvalidIDsBeforeRepositoryAccess(t *testing.T) {
	service := NewService(nil, nil, nil, nil)

	if _, err := service.Open(context.Background(), "not-a-uuid", CreateRoomInput{Kind: GameRoom}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Open() error = %v, want forbidden", err)
	}
	if err := service.Authorize(context.Background(), "not-a-uuid", "also-invalid"); !errors.Is(err, ErrForbidden) {
		t.Fatalf("Authorize() error = %v, want forbidden", err)
	}
}
