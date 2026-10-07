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
	game           Game
	saveMoveErr    error
	saveMoveGame   *Game
	terminalErr    error
	terminalGame   *Game
	expiredGames   []Game
	completionJobs []string
	completedJobs  []string
	retriedJobs    []string
	claimGame      *Game
	claimErr       error
	claimReason    string
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
	if r.saveMoveErr != nil {
		if r.saveMoveGame != nil {
			r.game = *r.saveMoveGame
		}
		return r.game, r.saveMoveErr
	}
	if expected != r.game.CurrentFEN {
		return Game{}, ErrMoveConflict
	}
	r.game.CurrentFEN = update.FEN
	r.game.Status = update.Status
	r.game.Result = update.Result
	r.game.EndReason = update.Reason
	r.game.DrawOfferBy = ""
	update.Record.PlayerID = player
	r.game.Moves = append(r.game.Moves, update.Record)
	return r.game, nil
}

func (r *memoryRepository) Resign(_ context.Context, _, player string) (Game, error) {
	if r.terminalErr != nil {
		return *r.terminalGame, r.terminalErr
	}
	if player != r.game.WhitePlayerID && player != r.game.BlackPlayerID {
		return Game{}, ErrNotFound
	}
	r.game.Status = StatusFinished
	r.game.Result = ResultBlackWin
	r.game.EndReason = "resignation"
	return r.game, nil
}

func (r *memoryRepository) OfferDraw(context.Context, string, string) (Game, error) {
	if r.terminalErr != nil {
		return *r.terminalGame, r.terminalErr
	}
	return r.game, nil
}

func (r *memoryRepository) AcceptDraw(context.Context, string, string) (Game, error) {
	if r.terminalErr != nil {
		return *r.terminalGame, r.terminalErr
	}
	r.game.Status = StatusFinished
	r.game.Result = ResultDraw
	return r.game, nil
}

func (r *memoryRepository) ClaimDraw(_ context.Context, _, _, reason string) (Game, error) {
	r.claimReason = reason
	if r.claimErr != nil {
		return *r.claimGame, r.claimErr
	}
	if r.claimGame != nil {
		return *r.claimGame, nil
	}
	return Game{}, ErrDrawUnavailable
}

func (r *memoryRepository) ExpireDueGames(context.Context, int) ([]Game, error) {
	return r.expiredGames, nil
}

func (r *memoryRepository) ClaimCompletionJobs(context.Context, int) ([]string, error) {
	jobs := append([]string(nil), r.completionJobs...)
	r.completionJobs = nil
	return jobs, nil
}

func (r *memoryRepository) CompleteCompletionJob(_ context.Context, gameID string) error {
	r.completedJobs = append(r.completedJobs, gameID)
	return nil
}

func (r *memoryRepository) RetryCompletionJob(_ context.Context, gameID string, _ error) error {
	r.retriedJobs = append(r.retriedJobs, gameID)
	return nil
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
	err     error
}

func (r *ratingRecorder) RecordGame(_ context.Context, result ratings.GameResult) error {
	r.results = append(r.results, result)
	return r.err
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

func TestOpponentMoveClearsPendingDrawOffer(t *testing.T) {
	repository := &memoryRepository{game: Game{
		ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
		CurrentFEN: "rnbqkbnr/pppppppp/8/8/4P3/8/PPPP1PPP/RNBQKBNR b KQkq - 0 1",
		Status:     StatusActive, Result: ResultOngoing, DrawOfferBy: "white",
	}}
	service := NewService(repository, nil)
	game, err := service.SubmitMove(context.Background(), testGameID, "black", MoveInput{UCI: "e7e5"})
	if err != nil {
		t.Fatalf("black move error = %v", err)
	}
	if game.DrawOfferBy != "" {
		t.Fatalf("draw offer remains from %q after opponent move", game.DrawOfferBy)
	}
}

func TestFivefoldRepetitionEndsGame(t *testing.T) {
	start := chess.StartingPosition()
	history := []MoveRecord{
		{FENAfter: start.FEN()},
		{FENAfter: start.FEN()},
		{FENAfter: start.FEN()},
	}
	repeated, err := fivefoldRepetition(history, start)
	if err != nil || !repeated {
		t.Fatalf("fivefold repetition = %v, %v; want true", repeated, err)
	}
	if status, result, reason := determineResult(mustParsePosition(t, "4k3/8/8/8/8/8/8/R3K3 w - - 150 1")); status != StatusFinished || result != ResultDraw || reason != "seventyfive_move_rule" {
		t.Fatalf("75-move result = (%q, %q, %q)", status, result, reason)
	}
}

func TestClaimDrawPublishesCompletion(t *testing.T) {
	claimed := Game{
		ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
		Status: StatusFinished, Result: ResultDraw, EndReason: "threefold_repetition_claim",
		Mode: "blitz", Rated: true,
	}
	repository := &memoryRepository{claimGame: &claimed}
	publisher := new(eventRecorder)
	ratingUpdates := new(ratingRecorder)
	service := NewServiceWithRatings(repository, publisher, ratingUpdates)
	game, err := service.ClaimDraw(context.Background(), testGameID, "white", "threefold_repetition")
	if err != nil {
		t.Fatalf("ClaimDraw() error = %v", err)
	}
	if game.EndReason != "threefold_repetition_claim" || repository.claimReason != "threefold_repetition" {
		t.Fatalf("claimed game = %+v; repository reason = %q", game, repository.claimReason)
	}
	if len(publisher.events) != 1 || publisher.events[0].Type != "game_ended" || len(ratingUpdates.results) != 1 {
		t.Fatalf("claim side effects = events %+v, ratings %+v", publisher.events, ratingUpdates.results)
	}
}

func mustParsePosition(t *testing.T, fen string) chess.Position {
	t.Helper()
	position, err := chess.ParseFEN(fen)
	if err != nil {
		t.Fatal(err)
	}
	return position
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

func TestClockRemainingUsesPositionTurnAndClampsFutureTimestamp(t *testing.T) {
	now := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	updatedAt := now.Add(-1500 * time.Millisecond)
	game := Game{
		CurrentFEN: chess.StartingPosition().FEN(), WhiteClockMS: 4000, BlackClockMS: 9000,
		ClockUpdatedAt: &updatedAt,
	}
	color, remaining, err := clockRemaining(game, now)
	if err != nil || color != chess.White || remaining != 2500 {
		t.Fatalf("white clock = (%v, %d, %v), want white/2500ms", color, remaining, err)
	}
	game.CurrentFEN = "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR b KQkq - 0 1"
	color, remaining, err = clockRemaining(game, now)
	if err != nil || color != chess.Black || remaining != 7500 {
		t.Fatalf("black clock = (%v, %d, %v), want black/7500ms", color, remaining, err)
	}
	future := now.Add(time.Second)
	game.ClockUpdatedAt = &future
	_, remaining, err = clockRemaining(game, now)
	if err != nil || remaining != game.BlackClockMS {
		t.Fatalf("future clock timestamp remaining = %d, %v; want %dms", remaining, err, game.BlackClockMS)
	}
}

func TestExpiredMovePublishesGameEndWithoutApplyingMove(t *testing.T) {
	start := chess.StartingPosition().FEN()
	finished := Game{
		ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
		CurrentFEN: start, Status: StatusFinished, Result: ResultBlackWin,
		EndReason: "timeout", Mode: "blitz", Rated: true,
	}
	repository := &memoryRepository{
		game: Game{
			ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
			CurrentFEN: start, Status: StatusActive, Result: ResultOngoing,
		},
		saveMoveErr: ErrTimeExpired, saveMoveGame: &finished,
	}
	publisher := new(eventRecorder)
	service := NewService(repository, publisher)
	game, err := service.SubmitMove(context.Background(), testGameID, "white", MoveInput{UCI: "e2e4"})
	if !errors.Is(err, ErrTimeExpired) {
		t.Fatalf("late move error = %v, want ErrTimeExpired", err)
	}
	if game.Status != StatusFinished || game.CurrentFEN != start {
		t.Fatalf("late move changed game position or missed timeout: %+v", game)
	}
	if len(publisher.events) != 1 || publisher.events[0].Type != "game_ended" {
		t.Fatalf("timeout events = %+v, want one game_ended event", publisher.events)
	}
}

func TestTerminalActionsHonorExpiredClock(t *testing.T) {
	start := chess.StartingPosition().FEN()
	finished := Game{
		ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
		CurrentFEN: start, Status: StatusFinished, Result: ResultBlackWin,
		EndReason: "timeout", Mode: "blitz", Rated: true,
	}
	actions := []struct {
		name string
		run  func(*Service) (Game, error)
	}{
		{name: "resign", run: func(service *Service) (Game, error) { return service.Resign(context.Background(), testGameID, "white") }},
		{name: "offer draw", run: func(service *Service) (Game, error) {
			return service.OfferDraw(context.Background(), testGameID, "white")
		}},
		{name: "accept draw", run: func(service *Service) (Game, error) {
			return service.AcceptDraw(context.Background(), testGameID, "white")
		}},
	}
	for _, action := range actions {
		t.Run(action.name, func(t *testing.T) {
			repository := &memoryRepository{
				game: Game{
					ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
					CurrentFEN: start, Status: StatusActive, Result: ResultOngoing,
				},
				terminalErr: ErrTimeExpired, terminalGame: &finished,
			}
			publisher := new(eventRecorder)
			service := NewService(repository, publisher)
			game, err := action.run(service)
			if !errors.Is(err, ErrTimeExpired) {
				t.Fatalf("action error = %v, want ErrTimeExpired", err)
			}
			if game.Result != ResultBlackWin || game.EndReason != "timeout" {
				t.Fatalf("action result = %+v, want timeout loss", game)
			}
			if len(publisher.events) != 1 || publisher.events[0].Type != "game_ended" {
				t.Fatalf("timeout events = %+v, want one game_ended event", publisher.events)
			}
		})
	}
}

func TestExpireDueGamesPublishesAndRecordsCompletion(t *testing.T) {
	finished := Game{
		ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
		Status: StatusFinished, Result: ResultBlackWin, EndReason: "timeout",
		Mode: "blitz", Rated: true,
	}
	repository := &memoryRepository{expiredGames: []Game{finished}}
	publisher := new(eventRecorder)
	ratingUpdates := new(ratingRecorder)
	service := NewServiceWithRatings(repository, publisher, ratingUpdates)
	if err := service.ExpireDueGames(context.Background(), 10); err != nil {
		t.Fatalf("ExpireDueGames() error = %v", err)
	}
	if len(publisher.events) != 1 || publisher.events[0].Type != "game_ended" {
		t.Fatalf("timeout events = %+v, want one game_ended event", publisher.events)
	}
	if len(ratingUpdates.results) != 1 || ratingUpdates.results[0].GameID != finished.ID {
		t.Fatalf("rating updates = %+v, want timed-out game result", ratingUpdates.results)
	}
}

func TestCompletionOutboxAcknowledgesSuccessAndRetriesFailure(t *testing.T) {
	game := Game{
		ID: testGameID, WhitePlayerID: "white", BlackPlayerID: "black",
		Status: StatusFinished, Result: ResultWhiteWin, Mode: "blitz", Rated: true,
	}
	repository := &memoryRepository{game: game, completionJobs: []string{testGameID}}
	ratings := &ratingRecorder{}
	service := NewServiceWithRatings(repository, nil, ratings)
	if err := service.ProcessCompletionJobs(context.Background(), 10); err != nil {
		t.Fatalf("ProcessCompletionJobs() error = %v", err)
	}
	if len(ratings.results) != 1 || len(repository.completedJobs) != 1 || repository.completedJobs[0] != testGameID {
		t.Fatalf("success processing = ratings %+v, acknowledgements %+v", ratings.results, repository.completedJobs)
	}

	repository.completionJobs = []string{testGameID}
	ratings.err = errors.New("temporary ratings database failure")
	if err := service.ProcessCompletionJobs(context.Background(), 10); err == nil {
		t.Fatal("ProcessCompletionJobs() error = nil, want retriable failure")
	}
	if len(repository.retriedJobs) != 1 || repository.retriedJobs[0] != testGameID {
		t.Fatalf("retry requests = %v, want game %s", repository.retriedJobs, testGameID)
	}
	if len(repository.completedJobs) != 1 {
		t.Fatalf("failed job was acknowledged: %v", repository.completedJobs)
	}
}
