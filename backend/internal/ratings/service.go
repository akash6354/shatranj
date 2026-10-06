package ratings

import (
	"context"
	"fmt"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Get(ctx context.Context, userID string, mode Mode) (Record, error) {
	if !mode.Valid() {
		return Record{}, ErrUnsupportedMode
	}
	return s.repository.Get(ctx, userID, mode)
}

func (s *Service) Preview(mode Mode, whiteRating, blackRating, whiteGames, blackGames int, result string) (Change, Change, error) {
	if !mode.Valid() {
		return Change{}, Change{}, ErrUnsupportedMode
	}
	whiteScore, err := ScoreForResult(result, true)
	if err != nil {
		return Change{}, Change{}, err
	}
	whiteNew, err := CalculateEloWithK(whiteRating, blackRating, whiteScore, kFactor(mode, whiteGames))
	if err != nil {
		return Change{}, Change{}, err
	}
	blackNew, err := CalculateEloWithK(blackRating, whiteRating, 1-whiteScore, kFactor(mode, blackGames))
	if err != nil {
		return Change{}, Change{}, err
	}
	return Change{Mode: mode, OldRating: whiteRating, NewRating: whiteNew, RatingDelta: whiteNew - whiteRating},
		Change{Mode: mode, OldRating: blackRating, NewRating: blackNew, RatingDelta: blackNew - blackRating}, nil
}

// RecordGame applies one final rated result. Repository implementations must
// make this operation idempotent on GameID.
func (s *Service) RecordGame(ctx context.Context, result GameResult) error {
	if result.GameID == "" || result.WhiteUserID == "" || result.BlackUserID == "" || result.WhiteUserID == result.BlackUserID {
		return fmt.Errorf("%w: game and distinct player IDs are required", ErrInvalidResult)
	}
	if !result.Mode.Valid() {
		return ErrUnsupportedMode
	}
	if _, err := ScoreForResult(result.Result, true); err != nil {
		return err
	}
	return s.repository.RecordGame(ctx, result)
}
