package puzzles

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

const testPuzzleID = "11111111-1111-4111-8111-111111111111"
const testUserID = "22222222-2222-4222-8222-222222222222"

type fakeRepository struct {
	puzzle  Puzzle
	attempt AttemptInput
}

func (r *fakeRepository) List(context.Context, Filter) ([]Puzzle, error) {
	return []Puzzle{r.puzzle}, nil
}

func (r *fakeRepository) Get(context.Context, string) (Puzzle, error) {
	return r.puzzle, nil
}

func (r *fakeRepository) Daily(context.Context, time.Time) (Puzzle, error) {
	return r.puzzle, nil
}

func (r *fakeRepository) Rating(context.Context, string) (int, error) {
	return defaultPuzzleRating, nil
}

func (r *fakeRepository) RecordAttempt(_ context.Context, attempt AttemptInput) (Attempt, error) {
	r.attempt = attempt
	return Attempt{PuzzleID: attempt.PuzzleID, Correct: attempt.Correct}, nil
}

func (r *fakeRepository) StartRush(context.Context, string, int) (RushSession, error) {
	return RushSession{}, nil
}

func (r *fakeRepository) GetRush(context.Context, string, string) (RushSession, error) {
	return RushSession{}, nil
}

func (r *fakeRepository) SubmitRushAnswer(context.Context, string, string, string, []string, bool) (RushSession, Attempt, error) {
	return RushSession{}, Attempt{}, nil
}

func TestSubmitPuzzleSolution(t *testing.T) {
	repository := &fakeRepository{puzzle: Puzzle{
		ID:       testPuzzleID,
		FEN:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Solution: []string{"e2e4", "e7e5"},
		Status:   StatusPublished,
	}}
	service := NewService(repository)

	attempt, err := service.Submit(context.Background(), testUserID, testPuzzleID, []string{"e2e4", "e7e5"})
	if err != nil {
		t.Fatal(err)
	}
	if !attempt.Correct || !repository.attempt.Correct {
		t.Fatal("expected correct solution to be recorded")
	}

	attempt, err = service.Submit(context.Background(), testUserID, testPuzzleID, []string{"e2e4", "c7c5"})
	if err != nil {
		t.Fatal(err)
	}
	if attempt.Correct || repository.attempt.Correct {
		t.Fatal("expected legal but incorrect line to be recorded as incorrect")
	}

	attempt, err = service.Submit(context.Background(), testUserID, testPuzzleID, []string{"e2e5"})
	if err != nil {
		t.Fatal(err)
	}
	if attempt.Correct {
		t.Fatal("illegal answer was marked correct")
	}
}

func TestPuzzleSolutionIsNotSerialized(t *testing.T) {
	body, err := json.Marshal(Puzzle{ID: testPuzzleID, Solution: []string{"e2e4"}})
	if err != nil {
		t.Fatal(err)
	}
	if string(body) == "" || strings.Contains(string(body), "e2e4") || strings.Contains(string(body), "solution") {
		t.Fatalf("solution leaked in puzzle response: %s", body)
	}
}

func TestStoredPuzzleRejectsIllegalSolution(t *testing.T) {
	err := validateStoredPuzzle(Puzzle{
		FEN:      "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Solution: []string{"e2e5"},
	})
	if err == nil {
		t.Fatal("illegal stored solution accepted")
	}
}

func TestPuzzleRatingUpdate(t *testing.T) {
	userAfter, puzzleAfter, err := updatedPuzzleRatings(1200, 1200, true)
	if err != nil {
		t.Fatal(err)
	}
	if userAfter <= 1200 || puzzleAfter >= 1200 {
		t.Fatalf("successful solve ratings = user:%d puzzle:%d", userAfter, puzzleAfter)
	}
	userAfter, puzzleAfter, err = updatedPuzzleRatings(1200, 1200, false)
	if err != nil {
		t.Fatal(err)
	}
	if userAfter >= 1200 || puzzleAfter <= 1200 {
		t.Fatalf("failed solve ratings = user:%d puzzle:%d", userAfter, puzzleAfter)
	}
}
