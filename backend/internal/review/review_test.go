package review

import (
	"context"
	"testing"

	"github.com/shatranj/backend/internal/analysis"
	"github.com/shatranj/backend/internal/games"
)

const reviewGameID = "33333333-3333-4333-8333-333333333333"

type fakeGameReader struct {
	game games.Game
}

func (r fakeGameReader) Get(context.Context, string) (games.Game, error) {
	return r.game, nil
}

type fakeRepository struct {
	review    Review
	available bool
}

func (r *fakeRepository) Request(_ context.Context, gameID, _ string, available bool) (Review, error) {
	r.available = available
	r.review.GameID = gameID
	if available {
		r.review.Status = StatusPending
	} else {
		r.review.Status = StatusUnavailable
	}
	return r.review, nil
}
func (r *fakeRepository) Get(context.Context, string) (Review, error) {
	return r.review, nil
}

func TestReviewRequestAuthorizationAndAvailability(t *testing.T) {
	repository := new(fakeRepository)
	game := games.Game{ID: reviewGameID, WhitePlayerID: "white", BlackPlayerID: "black", Status: games.StatusFinished}
	service := NewService(repository, fakeGameReader{game}, false)
	got, err := service.Request(context.Background(), reviewGameID, "white")
	if err != nil || got.Status != StatusUnavailable || repository.available {
		t.Fatalf("unavailable request = %+v, %v", got, err)
	}
	if _, err := service.Request(context.Background(), reviewGameID, "other"); err != ErrForbidden {
		t.Fatalf("unauthorized request error = %v", err)
	}

	game.Status = games.StatusActive
	service = NewService(repository, fakeGameReader{game}, true)
	if _, err := service.Request(context.Background(), reviewGameID, "white"); err != ErrGameNotFinished {
		t.Fatalf("unfinished game request error = %v", err)
	}
}

func TestReviewReturnsUnavailableMessageAndResult(t *testing.T) {
	repository := &fakeRepository{review: Review{
		GameID: reviewGameID, Status: StatusUnavailable,
	}}
	service := NewService(repository, fakeGameReader{games.Game{
		ID: reviewGameID, WhitePlayerID: "white", BlackPlayerID: "black",
	}}, false)
	got, err := service.Get(context.Background(), reviewGameID, "black")
	if err != nil || got.Message == "" {
		t.Fatalf("unavailable review = %+v, %v", got, err)
	}

	repository.review = Review{
		GameID: reviewGameID, Status: StatusCompleted,
		Result: &analysis.Result{GameID: reviewGameID, WhiteAccuracy: 98.5},
	}
	got, err = service.Get(context.Background(), reviewGameID, "black")
	if err != nil || got.Result == nil || got.Result.WhiteAccuracy != 98.5 {
		t.Fatalf("completed review = %+v, %v", got, err)
	}
}

func TestAccuracyAndMoveComparison(t *testing.T) {
	if got := AccuracyFromLoss(0); got != 100 {
		t.Fatalf("zero-loss accuracy = %v", got)
	}
	comparison := CompareMoves(50, 20, analysis.White, "e2e4", "e2e4")
	if !comparison.BestMovePlayed || comparison.CentipawnLoss != 30 {
		t.Fatalf("comparison = %+v", comparison)
	}
}
