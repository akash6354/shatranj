package analysis

import (
	"context"
	"testing"

	"github.com/akash6354/shatranj/backend/internal/chess"
)

type fakeEvaluator map[string]Evaluation

func (f fakeEvaluator) Evaluate(_ context.Context, fen string) (Evaluation, error) {
	return f[fen], nil
}

type fakeStore struct {
	result Result
	err    error
}

func (s *fakeStore) Claim(context.Context) (Task, error) { return Task{}, ErrNoJob }
func (s *fakeStore) Complete(_ context.Context, result Result) error {
	s.result = result
	return s.err
}
func (s *fakeStore) Fail(context.Context, string, error) error     { return nil }
func (s *fakeStore) MarkUnavailable(context.Context, string) error { return nil }

func TestWorkerAnalyzesStoredGameMoves(t *testing.T) {
	start := chess.StartingPosition()
	whiteMove, err := chess.ParseMove("e2e4")
	if err != nil {
		t.Fatal(err)
	}
	afterWhite, err := start.MakeMove(whiteMove)
	if err != nil {
		t.Fatal(err)
	}
	blackMove, err := chess.ParseMove("e7e5")
	if err != nil {
		t.Fatal(err)
	}
	afterBlack, err := afterWhite.MakeMove(blackMove)
	if err != nil {
		t.Fatal(err)
	}
	evaluator := fakeEvaluator{
		start.FEN():      {ScoreCP: 0, BestMove: "e2e4"},
		afterWhite.FEN(): {ScoreCP: 35, BestMove: "e7e5"},
		afterBlack.FEN(): {ScoreCP: 55, BestMove: "g1f3"},
	}
	store := new(fakeStore)
	worker := NewWorker(store, evaluator, nil)
	task := Task{
		ReviewID: "review", GameID: "game", WhiteID: "white", BlackID: "black",
		Moves: []PlayedMove{
			{PlayerID: "white", UCI: "e2e4", SAN: "e4", FENAfter: afterWhite.FEN()},
			{PlayerID: "black", UCI: "e7e5", SAN: "e5", FENAfter: afterBlack.FEN()},
		},
	}
	if err := worker.process(context.Background(), task); err != nil {
		t.Fatal(err)
	}
	if store.result.GameID != "game" || len(store.result.Moves) != 2 {
		t.Fatalf("analysis result = %+v", store.result)
	}
	if store.result.Moves[0].Classification != string(Best) || store.result.Moves[1].CentipawnLoss != 20 {
		t.Fatalf("move analysis = %+v", store.result.Moves)
	}
	if store.result.WhiteAccuracy != 100 || store.result.BlackAccuracy != moveAccuracy(20) {
		t.Fatalf("accuracies = white %.2f black %.2f", store.result.WhiteAccuracy, store.result.BlackAccuracy)
	}
}

func TestWorkerRejectsCorruptStoredHistory(t *testing.T) {
	worker := NewWorker(new(fakeStore), fakeEvaluator{}, nil)
	err := worker.process(context.Background(), Task{
		WhiteID: "white", BlackID: "black",
		Moves: []PlayedMove{{PlayerID: "black", UCI: "e2e4", FENAfter: "corrupt"}},
	})
	if err == nil {
		t.Fatal("corrupt move history was analyzed")
	}
}

func TestMoveClassificationAndAccuracy(t *testing.T) {
	if got := classifyMove(0, true, false, 0); got != Best {
		t.Fatalf("zero-loss classification = %s", got)
	}
	if got := classifyMove(250, false, false, 0); got != Blunder {
		t.Fatalf("large-loss classification = %s", got)
	}
	if got := classifyMove(0, true, true, 200); got != Brilliant {
		t.Fatalf("only-legal winning move classification = %s", got)
	}
	if got := AccuracyFromLoss(-10); got != 100 {
		t.Fatalf("negative-loss accuracy = %v", got)
	}
}
