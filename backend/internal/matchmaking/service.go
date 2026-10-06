package matchmaking

import (
	"context"
	"crypto/rand"
	"fmt"

	"github.com/shatranj/backend/internal/games"
	"github.com/shatranj/backend/internal/ratings"
)

type Service struct {
	queue   QueueRepository
	ratings RatingReader
	games   GameCreator
}

func NewService(queue QueueRepository, ratingReader RatingReader, gameCreator GameCreator) *Service {
	return &Service{queue: queue, ratings: ratingReader, games: gameCreator}
}

func (s *Service) Join(ctx context.Context, userID string, request Request) (Entry, error) {
	if userID == "" || !request.Mode.Valid() || request.Mode == ratings.ModePuzzle {
		return Entry{}, fmt.Errorf("%w: valid user and rated game mode are required", ErrInvalidQueue)
	}
	if request.TimeControl.InitialSeconds < 1 || request.TimeControl.InitialSeconds > 86400 ||
		request.TimeControl.IncrementSeconds < 0 || request.TimeControl.IncrementSeconds > 3600 {
		return Entry{}, fmt.Errorf("%w: invalid time control", ErrInvalidQueue)
	}
	rating, err := s.ratings.Get(ctx, userID, request.Mode)
	if err != nil {
		return Entry{}, fmt.Errorf("load matchmaking rating: %w", err)
	}
	entry := Entry{
		UserID: userID, Mode: request.Mode, TimeControl: request.TimeControl,
		Rating: rating.Rating, Status: StatusQueued,
	}
	joined, opponent, err := s.queue.JoinQueue(ctx, entry)
	if err != nil {
		return Entry{}, err
	}
	if opponent == nil {
		return joined, nil
	}
	return s.createMatch(ctx, joined, *opponent)
}

func (s *Service) createMatch(ctx context.Context, joining, opponent Entry) (Entry, error) {
	var colorChoice [1]byte
	if _, err := rand.Read(colorChoice[:]); err != nil {
		if abortErr := s.queue.AbortMatch(ctx, joining.MatchID); abortErr != nil {
			return Entry{}, fmt.Errorf("choose match colors: %w (abort matchmaking pair: %v)", err, abortErr)
		}
		return Entry{}, fmt.Errorf("choose match colors: %w", err)
	}
	white, black := opponent, joining
	if colorChoice[0]&1 == 1 {
		white, black = joining, opponent
	}
	input := games.CreateInput{
		TimeControl: joining.TimeControl,
		Mode:        string(joining.Mode),
		Rated:       true,
	}
	game, err := s.games.Create(ctx, white.UserID, input)
	if err != nil {
		if abortErr := s.queue.AbortMatch(ctx, joining.MatchID); abortErr != nil {
			return Entry{}, fmt.Errorf("create matched game: %w (abort matchmaking pair: %v)", err, abortErr)
		}
		return Entry{}, fmt.Errorf("create matched game: %w", err)
	}
	if _, err := s.games.Join(ctx, game.ID, black.UserID); err != nil {
		if abortErr := s.queue.AbortMatch(ctx, joining.MatchID); abortErr != nil {
			return Entry{}, fmt.Errorf("join matched game: %w (abort matchmaking pair: %v)", err, abortErr)
		}
		return Entry{}, fmt.Errorf("join matched game: %w", err)
	}
	if err := s.queue.CompleteMatch(ctx, joining.MatchID, game.ID); err != nil {
		return Entry{}, fmt.Errorf("persist matchmaking result: %w", err)
	}
	joined, err := s.queue.GetQueueStatus(ctx, joining.UserID)
	if err != nil {
		return Entry{}, fmt.Errorf("load completed matchmaking status: %w", err)
	}
	return joined, nil
}

func (s *Service) Leave(ctx context.Context, userID string) error {
	if userID == "" {
		return ErrQueueNotFound
	}
	return s.queue.LeaveQueue(ctx, userID)
}

func (s *Service) Status(ctx context.Context, userID string) (Entry, error) {
	if userID == "" {
		return Entry{}, ErrQueueNotFound
	}
	return s.queue.GetQueueStatus(ctx, userID)
}
