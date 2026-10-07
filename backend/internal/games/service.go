package games

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/shatranj/backend/internal/achievements"
	"github.com/shatranj/backend/internal/chess"
	"github.com/shatranj/backend/internal/ratings"
)

type Publisher interface {
	Publish(gameID string, event any)
}

type Service struct {
	repository   Repository
	publisher    Publisher
	ratings      RatingUpdater
	achievements AchievementTracker
}

type AchievementTracker interface {
	HandleEvent(context.Context, achievements.Event) ([]achievements.Achievement, error)
}

type RatingUpdater interface {
	RecordGame(context.Context, ratings.GameResult) error
}

func NewService(repository Repository, publisher Publisher) *Service {
	return &Service{repository: repository, publisher: publisher}
}

func NewServiceWithRatings(repository Repository, publisher Publisher, ratingUpdater RatingUpdater) *Service {
	return &Service{repository: repository, publisher: publisher, ratings: ratingUpdater}
}

func (s *Service) SetAchievementTracker(tracker AchievementTracker) {
	s.achievements = tracker
}

func (s *Service) Create(ctx context.Context, userID string, input CreateInput) (Game, error) {
	if userID == "" {
		return Game{}, ErrForbidden
	}
	if input.TimeControl.InitialSeconds == 0 {
		input.TimeControl.InitialSeconds = 300
	}
	if input.Mode == "" {
		input.Mode = "blitz"
	}
	if input.Mode != "bullet" && input.Mode != "blitz" && input.Mode != "rapid" {
		return Game{}, fmt.Errorf("%w: unsupported game mode", ErrInvalidRequest)
	}
	if input.TimeControl.InitialSeconds < 0 || input.TimeControl.InitialSeconds > 86400 ||
		input.TimeControl.IncrementSeconds < 0 || input.TimeControl.IncrementSeconds > 3600 {
		return Game{}, fmt.Errorf("%w: invalid time control", ErrInvalidRequest)
	}
	game, err := s.repository.Create(ctx, userID, input)
	if err != nil {
		return Game{}, err
	}
	s.publish(game.ID, GameEvent{Type: "game_created", GameID: game.ID, Game: &game, CreatedAt: time.Now().UTC()})
	return game, nil
}

func (s *Service) Join(ctx context.Context, gameID, userID string) (Game, error) {
	if userID == "" {
		return Game{}, ErrForbidden
	}
	if !validGameID(gameID) {
		return Game{}, ErrNotFound
	}
	game, err := s.repository.Join(ctx, gameID, userID)
	if err != nil {
		return Game{}, err
	}
	s.publish(game.ID, GameEvent{Type: "game_joined", GameID: game.ID, Game: &game, PlayerID: userID, Color: chess.Black, CreatedAt: time.Now().UTC()})
	return game, nil
}

func (s *Service) Get(ctx context.Context, gameID, userID string) (Game, error) {
	if !validGameID(gameID) {
		return Game{}, ErrNotFound
	}
	if userID == "" {
		return Game{}, ErrForbidden
	}
	game, err := s.repository.Get(ctx, gameID)
	if err != nil {
		return Game{}, err
	}
	if _, err := playerColor(game, userID); err != nil {
		return Game{}, err
	}
	return game, nil
}

func (s *Service) SubmitMove(ctx context.Context, gameID, userID string, input MoveInput) (Game, error) {
	if !validGameID(gameID) {
		return Game{}, ErrNotFound
	}
	if userID == "" {
		return Game{}, ErrForbidden
	}
	game, err := s.repository.Get(ctx, gameID)
	if err != nil {
		return Game{}, err
	}
	if game.Status != StatusActive {
		return Game{}, ErrGameNotActive
	}
	color, err := playerColor(game, userID)
	if err != nil {
		return Game{}, err
	}
	position, err := positionFromGame(game)
	if err != nil {
		return Game{}, err
	}
	if position.SideToMove != color {
		return Game{}, ErrNotYourTurn
	}
	move, err := chess.ParseMove(input.UCI)
	if err != nil {
		return Game{}, fmt.Errorf("%w: invalid UCI move", ErrInvalidRequest)
	}
	san, err := position.SAN(move)
	if err != nil {
		return Game{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	next, err := position.MakeMove(move)
	if err != nil {
		return Game{}, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}
	status, result, reason := determineResult(next)
	if status == StatusActive {
		repeated, err := fivefoldRepetition(game.Moves, next)
		if err != nil {
			return Game{}, err
		}
		if repeated {
			status, result, reason = StatusFinished, ResultDraw, "fivefold_repetition"
		}
	}
	record := MoveRecord{
		Number: position.FullmoveNumber, PlayerID: userID, UCI: move.String(),
		SAN: san, FENAfter: next.FEN(),
	}
	updated, err := s.repository.SaveMove(ctx, gameID, userID, game.CurrentFEN, MoveUpdate{
		PlayerID: userID, Move: move, Record: record, FEN: next.FEN(),
		Status: status, Result: result, Reason: reason,
	})
	if errors.Is(err, ErrTimeExpired) {
		return s.handleTimeout(ctx, gameID, userID, updated)
	}
	if err != nil {
		return Game{}, err
	}
	moveEvent := GameEvent{
		Type: "move", GameID: gameID, Game: &updated, Move: &record,
		PlayerID: userID, Result: result, Reason: reason, CreatedAt: time.Now().UTC(),
	}
	s.publish(gameID, moveEvent)
	if status == StatusFinished {
		s.publish(gameID, GameEvent{
			Type: "game_ended", GameID: gameID, Game: &updated, Move: &record,
			PlayerID: userID, Result: result, Reason: reason, CreatedAt: moveEvent.CreatedAt,
		})
		if err := s.recordCompletion(ctx, updated); err != nil {
			return Game{}, err
		}
	} else if next.IsInCheck(next.SideToMove) {
		s.publish(gameID, GameEvent{
			Type: "game_state", GameID: gameID, Game: &updated, Move: &record,
			PlayerID: userID, Result: result, Reason: "check", CreatedAt: moveEvent.CreatedAt,
		})
	}
	return updated, nil
}

func fivefoldRepetition(history []MoveRecord, next chess.Position) (bool, error) {
	key := next.RepetitionKey()
	occurrences := 0
	count := func(position chess.Position) {
		if position.RepetitionKey() == key {
			occurrences++
		}
	}
	count(chess.StartingPosition())
	for _, record := range history {
		position, err := chess.ParseFEN(record.FENAfter)
		if err != nil {
			return false, fmt.Errorf("parse historical position for repetition: %w", err)
		}
		count(position)
	}
	count(next)
	return occurrences >= 5, nil
}

func (s *Service) Resign(ctx context.Context, gameID, userID string) (Game, error) {
	if _, err := s.Get(ctx, gameID, userID); err != nil {
		return Game{}, err
	}
	game, err := s.repository.Resign(ctx, gameID, userID)
	if errors.Is(err, ErrTimeExpired) {
		return s.handleTimeout(ctx, gameID, userID, game)
	}
	if err != nil {
		return Game{}, err
	}
	event := GameEvent{Type: "game_ended", GameID: gameID, Game: &game, PlayerID: userID, Result: game.Result, Reason: game.EndReason, CreatedAt: time.Now().UTC()}
	s.publish(gameID, GameEvent{Type: "resigned", GameID: gameID, PlayerID: userID, CreatedAt: event.CreatedAt})
	s.publish(gameID, event)
	if err := s.recordCompletion(ctx, game); err != nil {
		return Game{}, err
	}
	return game, nil
}

func (s *Service) OfferDraw(ctx context.Context, gameID, userID string) (Game, error) {
	if _, err := s.Get(ctx, gameID, userID); err != nil {
		return Game{}, err
	}
	game, err := s.repository.OfferDraw(ctx, gameID, userID)
	if errors.Is(err, ErrTimeExpired) {
		return s.handleTimeout(ctx, gameID, userID, game)
	}
	if err != nil {
		return Game{}, err
	}
	s.publish(gameID, GameEvent{Type: "draw_offered", GameID: gameID, PlayerID: userID, CreatedAt: time.Now().UTC()})
	return game, nil
}

func (s *Service) AcceptDraw(ctx context.Context, gameID, userID string) (Game, error) {
	if _, err := s.Get(ctx, gameID, userID); err != nil {
		return Game{}, err
	}
	game, err := s.repository.AcceptDraw(ctx, gameID, userID)
	if errors.Is(err, ErrTimeExpired) {
		return s.handleTimeout(ctx, gameID, userID, game)
	}
	if err != nil {
		return Game{}, err
	}
	s.publish(gameID, GameEvent{Type: "game_ended", GameID: gameID, Game: &game, PlayerID: userID, Result: game.Result, Reason: game.EndReason, CreatedAt: time.Now().UTC()})
	if err := s.recordCompletion(ctx, game); err != nil {
		return Game{}, err
	}
	return game, nil
}

func (s *Service) ClaimDraw(ctx context.Context, gameID, userID, reason string) (Game, error) {
	if !validGameID(gameID) {
		return Game{}, ErrNotFound
	}
	if userID == "" {
		return Game{}, ErrForbidden
	}
	game, err := s.repository.ClaimDraw(ctx, gameID, userID, reason)
	if errors.Is(err, ErrTimeExpired) {
		return s.handleTimeout(ctx, gameID, userID, game)
	}
	if err != nil {
		return Game{}, err
	}
	s.publish(gameID, GameEvent{
		Type: "game_ended", GameID: gameID, Game: &game, PlayerID: userID,
		Result: game.Result, Reason: game.EndReason, CreatedAt: game.UpdatedAt,
	})
	if err := s.recordCompletion(ctx, game); err != nil {
		return Game{}, err
	}
	return game, nil
}

func (s *Service) handleTimeout(ctx context.Context, gameID, userID string, game Game) (Game, error) {
	if err := s.recordTimedOutGame(ctx, gameID, userID, game); err != nil {
		return game, fmt.Errorf("record timeout completion: %w", err)
	}
	return game, ErrTimeExpired
}

func (s *Service) recordTimedOutGame(ctx context.Context, gameID, userID string, game Game) error {
	s.publish(gameID, GameEvent{
		Type: "game_ended", GameID: gameID, Game: &game, PlayerID: userID,
		Result: game.Result, Reason: game.EndReason, CreatedAt: game.UpdatedAt,
	})
	return s.recordCompletion(ctx, game)
}

func (s *Service) ExpireDueGames(ctx context.Context, limit int) error {
	games, err := s.repository.ExpireDueGames(ctx, limit)
	if err != nil {
		return err
	}
	var completionErrors []error
	for _, game := range games {
		if err := s.recordTimedOutGame(ctx, game.ID, "", game); err != nil {
			completionErrors = append(completionErrors, fmt.Errorf("complete timed-out game %s: %w", game.ID, err))
		}
	}
	return errors.Join(completionErrors...)
}

func (s *Service) ProcessCompletionJobs(ctx context.Context, limit int) error {
	gameIDs, err := s.repository.ClaimCompletionJobs(ctx, limit)
	if err != nil {
		return err
	}
	var completionErrors []error
	for _, gameID := range gameIDs {
		game, processErr := s.repository.Get(ctx, gameID)
		if processErr == nil {
			processErr = s.recordCompletion(ctx, game)
		}
		if processErr != nil {
			if retryErr := s.repository.RetryCompletionJob(ctx, gameID, processErr); retryErr != nil {
				completionErrors = append(completionErrors, fmt.Errorf("retry completion job %s: %w", gameID, retryErr))
			}
			completionErrors = append(completionErrors, fmt.Errorf("process completion job %s: %w", gameID, processErr))
			continue
		}
		if err := s.repository.CompleteCompletionJob(ctx, gameID); err != nil {
			completionErrors = append(completionErrors, fmt.Errorf("acknowledge completion job %s: %w", gameID, err))
		}
	}
	return errors.Join(completionErrors...)
}

func (s *Service) recordCompletion(ctx context.Context, game Game) error {
	ratingErr := s.recordRating(ctx, game)
	achievementErr := s.recordAchievements(ctx, game)
	if ratingErr != nil {
		ratingErr = fmt.Errorf("record completion rating: %w", ratingErr)
	}
	if achievementErr != nil {
		achievementErr = fmt.Errorf("record completion achievements: %w", achievementErr)
	}
	return errors.Join(ratingErr, achievementErr)
}

func (s *Service) recordAchievements(ctx context.Context, game Game) error {
	if s.achievements == nil || game.BlackPlayerID == "" {
		return nil
	}
	for _, userID := range []string{game.WhitePlayerID, game.BlackPlayerID} {
		if _, err := s.achievements.HandleEvent(ctx, achievements.Event{
			UserID: userID, Type: "game_completed", SourceID: "game:" + game.ID, Amount: 1,
		}); err != nil {
			return fmt.Errorf("record game achievement: %w", err)
		}
	}
	return nil
}

func (s *Service) recordRating(ctx context.Context, game Game) error {
	if !game.Rated || s.ratings == nil {
		return nil
	}
	if game.BlackPlayerID == "" {
		return fmt.Errorf("cannot rate unfinished game %s", game.ID)
	}
	if err := s.ratings.RecordGame(ctx, ratings.GameResult{
		GameID: game.ID, WhiteUserID: game.WhitePlayerID,
		BlackUserID: game.BlackPlayerID, Mode: ratings.Mode(game.Mode), Result: string(game.Result),
	}); err != nil {
		return fmt.Errorf("record game ratings: %w", err)
	}
	return nil
}

func (s *Service) publish(gameID string, event GameEvent) {
	if s.publisher != nil {
		s.publisher.Publish(gameID, event)
	}
}
