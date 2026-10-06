package review

import (
	"context"
	"errors"
	"fmt"

	"github.com/shatranj/backend/internal/games"
)

type Service struct {
	repository Repository
	gameReader GameReader
	available  bool
}

func NewService(repository Repository, gameReader GameReader, engineAvailable bool) *Service {
	return &Service{repository: repository, gameReader: gameReader, available: engineAvailable}
}

func (s *Service) Request(ctx context.Context, gameID, userID string) (Review, error) {
	game, err := s.authorizedGame(ctx, gameID, userID)
	if err != nil {
		return Review{}, err
	}
	if game.Status != games.StatusFinished {
		return Review{}, ErrGameNotFinished
	}
	return s.repository.Request(ctx, gameID, userID, s.available)
}

func (s *Service) Get(ctx context.Context, gameID, userID string) (Review, error) {
	if _, err := s.authorizedGame(ctx, gameID, userID); err != nil {
		return Review{}, err
	}
	review, err := s.repository.Get(ctx, gameID)
	if err != nil {
		return Review{}, err
	}
	switch review.Status {
	case StatusUnavailable:
		review.Message = "Stockfish analysis is not configured"
	case StatusFailed:
		review.Message = "Analysis failed; request a new review to retry"
	}
	return review, nil
}

func (s *Service) authorizedGame(ctx context.Context, gameID, userID string) (games.Game, error) {
	if userID == "" {
		return games.Game{}, ErrForbidden
	}
	if !validGameID(gameID) {
		return games.Game{}, ErrNotFound
	}
	game, err := s.gameReader.Get(ctx, gameID)
	if err != nil {
		if errors.Is(err, games.ErrNotFound) {
			return games.Game{}, ErrNotFound
		}
		return games.Game{}, fmt.Errorf("load game for review: %w", err)
	}
	if userID != game.WhitePlayerID && userID != game.BlackPlayerID {
		return games.Game{}, ErrForbidden
	}
	return game, nil
}

func validGameID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for index, value := range id {
		switch index {
		case 8, 13, 18, 23:
			if value != '-' {
				return false
			}
		default:
			if !(value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F') {
				return false
			}
		}
	}
	return true
}
