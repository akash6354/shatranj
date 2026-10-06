package achievements

import (
	"context"
	"errors"
	"testing"
)

func TestHandleEventRequiresIdempotencySource(t *testing.T) {
	service := NewService(nil)
	_, err := service.HandleEvent(context.Background(), Event{
		UserID: "00000000-0000-4000-8000-000000000001",
		Type:   "game_completed",
		Amount: 1,
	})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("HandleEvent() error = %v, want ErrInvalid", err)
	}
}
