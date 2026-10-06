package games

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/shatranj/backend/internal/chess"
	"github.com/shatranj/backend/internal/ratings"
)

const testGameID = "00000000-0000-4000-8000-000000000001"

type memoryRepository struct {
	game Game
}

func (r *memoryRepository) Create(_ context.Context, player string, input CreateInput) (Game, error) {
	r.game = Game{
		ID: testGameID, WhitePlayerID: player, TimeControl: input.TimeControl,
		CurrentFEN: "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1",
		Status:     StatusWaiting, Result: ResultOngoing, CreatedAt: time.Now(),
	}
	return r.game, nil
}

func (r *memoryRepository) Join(_ context.Context, _ string, player string) (Game, error) {
	if r.game.Status != StatusWaiting {
		return Game{}, ErrGameFull
	}
	r.game.BlackPlayerID = player
	r.game.Status = StatusActive
	return r.game, nil
}

func (r *memoryRepository) Get(_ context.Context, id string) (Game, error) {
	if id != r.game.ID {
		return Game{}, ErrNotFound
	}
	return r.game, nil
}

func (r *memoryRepository) SaveMove(_ context.Context, id, player, expected string, update MoveUpdate) (Game, error) {
	if id != r.game.ID {
		return Game{}, ErrNotFound
	}
	if expected != r.game.CurrentFEN {
		return Game{}, ErrMoveConflict
	}
	r.game.CurrentFEN = update.FEN
	r.game.Status = update.Status
	r.game.Result = update.Result
	r.game.EndReason = update.Reason
	update.Record.PlayerID = player
	r.game.Moves = append(r.game.Moves, update.Record)
	return r.game, nil
}

func (r *memoryRepository) Resign(_ context.Context, _, player string) (Game, error) {
	if player != r.game.WhitePlayerID && player != r.game.BlackPlayerID {
		return Game{}, ErrNotFound
	}
	r.game.Status = StatusFinished
	r.game.Result = ResultBlackWin
	r.game.EndReason = "resignation"
	return r.game, nil
}

func (r *memoryRepository) OfferDraw(context.Context, string, string) (Game, error) {
	return r.game, nil
}

func (r *memoryRepository) AcceptDraw(context.Context, string, string) (Game, error) {
	r.game.Status = StatusFinished
	r.game.Result = ResultDraw
	return r.game, nil
}

type eventRecorder struct {
	events []GameEvent
}

func (r *eventRecorder) Publish(_ string, event any) {
	if gameEvent, ok := event.(GameEvent); ok {
		r.events = append(r.events, gameEvent)
	}
}

type ratingRecorder struct {
	results []ratings.GameResult
}

func (r *ratingRecorder) RecordGame(_ context.Context, result ratings.GameResult) error {
	r.results = append(r.results, result)
	return nil
}

func TestCreateJoinAndMoveAuthorization(t *testing.T) {
	repository := new(memoryRepository)
	publisher := new(eventRecorder)
	service := NewService(repository, publisher)
	game, err := service.Create(context.Background(), "white", CreateInput{})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if game.Status != StatusWaiting || game.TimeControl.InitialSeconds != 300 {
		t.Fatalf("created game = %+v", game)
	}
	game, err = service.Join(context.Background(), game.ID, "black")
	if err != nil || game.Status != StatusActive {
		t.Fatalf("Join() = %+v, %v", game, err)
	}
	if _, err := service.SubmitMove(context.Background(), game.ID, "outsider", MoveInput{UCI: "e2e4"}); !errors.Is(err, ErrForbidden) {
		t.Fatalf("outsider move error = %v, want ErrForbidden", err)
	}
	if _, err := service.SubmitMove(context.Background(), game.ID, "black", MoveInput{UCI: "e7e5"}); !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf("out-of-turn move error = %v, want ErrNotYourTurn", err)
	}
	updated, err := service.SubmitMove(context.Background(), game.ID, "white", MoveInput{UCI: "e2e4"})
	if err != nil {
		t.Fatalf("white move error = %v", err)
	}
	if updated.CurrentFEN == game.CurrentFEN || len(updated.Moves) != 1 || updated.Moves[0].SAN != "e4" {
		t.Fatalf("move did not update game: %+v", updated)
	}
	if got := publisher.events[len(publisher.events)-1].Type; got != "move" {
		t.Fatalf("last published event = %q, want move", got)
	}
}

func TestCheckmateFinishesGame(t *testing.T) {
	ratingUpdates := new(ratingRecorder)
	repository := &memoryRepository{game: Game{
		ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
		CurrentFEN: "7k/8/5KQ1/8/8/8/8/8 w - - 0 1",
		Status:     StatusActive, Result: ResultOngoing, Mode: "blitz", Rated: true,
	}}
	service := NewServiceWithRatings(repository, nil, ratingUpdates)
	game, err := service.SubmitMove(context.Background(), testGameID, "white", MoveInput{UCI: "g6g7"})
	if err != nil {
		t.Fatalf("mating move error = %v", err)
	}
	if game.Status != StatusFinished || game.Result != ResultWhiteWin || game.EndReason != "checkmate" {
		t.Fatalf("mate result = %+v", game)
	}
	if len(ratingUpdates.results) != 1 || ratingUpdates.results[0].GameID != game.ID {
		t.Fatalf("rating updates = %+v", ratingUpdates.results)
	}
}

func TestTimeoutResultHonorsInsufficientMaterial(t *testing.T) {
	position, err := chess.ParseFEN("4k3/8/8/8/8/8/8/4K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if result, reason := ResultOnTimeout(position, chess.White); result != ResultDraw || reason != "timeout_insufficient_material" {
		t.Fatalf("bare-kings timeout = (%q, %q), want draw", result, reason)
	}
	position, err = chess.ParseFEN("4k3/8/8/8/8/8/8/R3K3 w - - 0 1")
	if err != nil {
		t.Fatal(err)
	}
	if result, reason := ResultOnTimeout(position, chess.White); result != ResultBlackWin || reason != "timeout" {
		t.Fatalf("timeout result = (%q, %q), want black win", result, reason)
	}
}
