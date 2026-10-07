package leaderboard

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shatranj/backend/internal/ratings"
)

var ErrInvalidRequest = errors.New("invalid leaderboard request")

const (
	defaultLimit = 50
	maxLimit     = 100
)

type Entry struct {
	Rank        int          `json:"rank"`
	UserID      string       `json:"user_id"`
	Username    string       `json:"username"`
	DisplayName string       `json:"display_name"`
	AvatarURL   string       `json:"avatar_url,omitempty"`
	Mode        ratings.Mode `json:"mode"`
	Rating      int          `json:"rating"`
	Highest     int          `json:"highest"`
	Games       int          `json:"games"`
	Wins        int          `json:"wins"`
	Draws       int          `json:"draws"`
	Losses      int          `json:"losses"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Top(ctx context.Context, mode ratings.Mode, limit int) ([]Entry, error) {
	if !mode.Valid() {
		return nil, fmt.Errorf("%w: unsupported rating mode", ErrInvalidRequest)
	}
	if limit <= 0 {
		limit = defaultLimit
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return s.repository.Top(ctx, mode, limit)
}
