package leaderboard

import (
	"context"
	"errors"
	"testing"

	"github.com/akash6354/shatranj/backend/internal/ratings"
)

type recordingRepository struct {
	gotMode  ratings.Mode
	gotLimit int
	entries  []Entry
	calls    int
}

func (r *recordingRepository) Top(_ context.Context, mode ratings.Mode, limit int) ([]Entry, error) {
	r.calls++
	r.gotMode = mode
	r.gotLimit = limit
	return r.entries, nil
}

func TestTopRejectsUnsupportedMode(t *testing.T) {
	repository := new(recordingRepository)
	service := NewService(repository)
	if _, err := service.Top(context.Background(), ratings.Mode("classical"), 10); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("Top() error = %v, want ErrInvalidRequest", err)
	}
	if repository.calls != 0 {
		t.Fatal("repository was queried for an unsupported mode")
	}
}

func TestTopClampsLimit(t *testing.T) {
	tests := []struct {
		name  string
		limit int
		want  int
	}{
		{name: "zero uses default", limit: 0, want: defaultLimit},
		{name: "negative uses default", limit: -5, want: defaultLimit},
		{name: "in range passes through", limit: 25, want: 25},
		{name: "above max is clamped", limit: 5000, want: maxLimit},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &recordingRepository{entries: []Entry{{Rank: 1, UserID: "u1"}}}
			service := NewService(repository)
			entries, err := service.Top(context.Background(), ratings.ModeBlitz, test.limit)
			if err != nil {
				t.Fatalf("Top() error = %v", err)
			}
			if repository.gotLimit != test.want {
				t.Fatalf("repository limit = %d, want %d", repository.gotLimit, test.want)
			}
			if repository.gotMode != ratings.ModeBlitz {
				t.Fatalf("repository mode = %s, want blitz", repository.gotMode)
			}
			if len(entries) != 1 {
				t.Fatalf("entries = %+v", entries)
			}
		})
	}
}
