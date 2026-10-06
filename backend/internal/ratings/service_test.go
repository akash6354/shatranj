package ratings

import (
	"context"
	"testing"
)

type memoryRatings struct {
	records map[string]Record
	games   []GameResult
}

func (r *memoryRatings) Get(_ context.Context, userID string, mode Mode) (Record, error) {
	if r.records == nil {
		r.records = make(map[string]Record)
	}
	record, ok := r.records[userID+string(mode)]
	if !ok {
		initial := DefaultRating
		if mode == ModePuzzle {
			initial = 1500
		}
		record = Record{UserID: userID, Mode: mode, Rating: initial}
		r.records[userID+string(mode)] = record
	}
	return record, nil
}

func (r *memoryRatings) RecordGame(_ context.Context, game GameResult) error {
	r.games = append(r.games, game)
	return nil
}

func TestRecordGameValidatesAndPersistsResult(t *testing.T) {
	repository := new(memoryRatings)
	service := NewService(repository)
	result := GameResult{
		GameID: "game-1", WhiteUserID: "white", BlackUserID: "black",
		Mode: ModeBlitz, Result: "1-0",
	}
	if err := service.RecordGame(context.Background(), result); err != nil {
		t.Fatalf("RecordGame() error = %v", err)
	}
	if len(repository.games) != 1 || repository.games[0] != result {
		t.Fatalf("persisted results = %+v", repository.games)
	}
	result.Result = "*"
	if err := service.RecordGame(context.Background(), result); err != ErrInvalidResult {
		t.Fatalf("ongoing result error = %v, want ErrInvalidResult", err)
	}
}

func TestPreviewRatingsAreZeroSum(t *testing.T) {
	service := NewService(new(memoryRatings))
	white, black, err := service.Preview(ModeBlitz, 1200, 1200, 0, 0, "1-0")
	if err != nil {
		t.Fatal(err)
	}
	if white.RatingDelta != 20 || black.RatingDelta != -20 {
		t.Fatalf("changes = %+v, %+v; want +20/-20 for provisional ratings", white, black)
	}
}
