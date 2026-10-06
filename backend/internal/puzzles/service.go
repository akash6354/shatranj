package puzzles

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/shatranj/backend/internal/achievements"
	"github.com/shatranj/backend/internal/chess"
)

const (
	defaultRushDuration = 180 * time.Second
	maxPuzzlePageSize   = 100
)

type Service struct {
	repository   Repository
	now          func() time.Time
	achievements AchievementTracker
}

type AchievementTracker interface {
	HandleEvent(context.Context, achievements.Event) ([]achievements.Achievement, error)
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository, now: time.Now}
}

func (s *Service) SetAchievementTracker(tracker AchievementTracker) {
	s.achievements = tracker
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Puzzle, error) {
	if filter.Limit < 0 {
		return nil, fmt.Errorf("%w: limit cannot be negative", ErrInvalidInput)
	}
	if filter.Limit == 0 {
		filter.Limit = 20
	} else if filter.Limit > maxPuzzlePageSize {
		filter.Limit = maxPuzzlePageSize
	}
	if filter.Offset < 0 {
		return nil, fmt.Errorf("%w: offset cannot be negative", ErrInvalidInput)
	}
	filter.Theme = normalizeTheme(filter.Theme)
	filter.Difficulty = strings.ToLower(strings.TrimSpace(filter.Difficulty))
	if filter.Difficulty != "" && filter.Difficulty != "easy" && filter.Difficulty != "medium" &&
		filter.Difficulty != "hard" && filter.Difficulty != "expert" {
		return nil, fmt.Errorf("%w: unsupported difficulty", ErrInvalidInput)
	}
	puzzles, err := s.repository.List(ctx, filter)
	if err != nil {
		return nil, err
	}
	for _, puzzle := range puzzles {
		if err := validateStoredPuzzle(puzzle); err != nil {
			return nil, err
		}
	}
	return puzzles, nil
}

func (s *Service) Get(ctx context.Context, id string) (Puzzle, error) {
	if !validID(id) {
		return Puzzle{}, ErrNotFound
	}
	puzzle, err := s.repository.Get(ctx, id)
	if err != nil {
		return Puzzle{}, err
	}
	if err := validateStoredPuzzle(puzzle); err != nil {
		return Puzzle{}, err
	}
	return puzzle, nil
}

func (s *Service) Daily(ctx context.Context) (Puzzle, error) {
	puzzle, err := s.repository.Daily(ctx, s.now())
	if err != nil {
		return Puzzle{}, err
	}
	if err := validateStoredPuzzle(puzzle); err != nil {
		return Puzzle{}, err
	}
	return puzzle, nil
}

func (s *Service) Rating(ctx context.Context, userID string) (int, error) {
	if userID == "" {
		return 0, ErrInvalidInput
	}
	return s.repository.Rating(ctx, userID)
}

func (s *Service) Submit(ctx context.Context, userID, puzzleID string, moves []string) (Attempt, error) {
	if userID == "" || !validID(puzzleID) || len(moves) == 0 || len(moves) > 100 {
		return Attempt{}, ErrInvalidInput
	}
	puzzle, err := s.Get(ctx, puzzleID)
	if err != nil {
		return Attempt{}, err
	}
	correct, err := solutionMatches(puzzle, moves)
	if err != nil {
		return Attempt{}, err
	}
	attempt, err := s.repository.RecordAttempt(ctx, AttemptInput{
		UserID: userID, PuzzleID: puzzleID, Moves: moves, Correct: correct,
	})
	if err != nil {
		return Attempt{}, err
	}
	if err := s.recordPuzzleAchievement(ctx, userID, attempt, correct); err != nil {
		return attempt, err
	}
	return attempt, nil
}

func (s *Service) StartRush(ctx context.Context, userID string, durationSeconds int) (RushSession, error) {
	if userID == "" {
		return RushSession{}, ErrInvalidInput
	}
	if durationSeconds == 0 {
		durationSeconds = int(defaultRushDuration.Seconds())
	}
	if !validRushDuration(durationSeconds) {
		return RushSession{}, fmt.Errorf("%w: rush time limit must be between 30 and 1800 seconds", ErrInvalidInput)
	}
	session, err := s.repository.StartRush(ctx, userID, durationSeconds)
	if err != nil {
		return RushSession{}, err
	}
	if session.CurrentPuzzle == nil {
		return RushSession{}, fmt.Errorf("%w: rush session has no puzzle", ErrInvalidPuzzle)
	}
	if err := validateStoredPuzzle(*session.CurrentPuzzle); err != nil {
		return RushSession{}, err
	}
	return session, nil
}

func (s *Service) GetRush(ctx context.Context, userID, sessionID string) (RushSession, error) {
	if userID == "" || !validID(sessionID) {
		return RushSession{}, ErrRushNotFound
	}
	session, err := s.repository.GetRush(ctx, userID, sessionID)
	if err != nil {
		return RushSession{}, err
	}
	if session.CurrentPuzzle != nil {
		if err := validateStoredPuzzle(*session.CurrentPuzzle); err != nil {
			return RushSession{}, err
		}
	}
	return session, nil
}

func (s *Service) SubmitRushAnswer(ctx context.Context, userID, sessionID string, answer RushAnswer) (RushSession, Attempt, error) {
	if userID == "" || !validID(sessionID) || !validID(answer.PuzzleID) || len(answer.Moves) == 0 || len(answer.Moves) > 100 {
		return RushSession{}, Attempt{}, ErrInvalidInput
	}
	session, err := s.repository.GetRush(ctx, userID, sessionID)
	if err != nil {
		return RushSession{}, Attempt{}, err
	}
	if session.Status != RushActive {
		return RushSession{}, Attempt{}, ErrRushFinished
	}
	if session.CurrentPuzzleID != answer.PuzzleID {
		return RushSession{}, Attempt{}, ErrRushPuzzle
	}
	puzzle, err := s.Get(ctx, answer.PuzzleID)
	if err != nil {
		return RushSession{}, Attempt{}, err
	}
	correct, err := solutionMatches(puzzle, answer.Moves)
	if err != nil {
		return RushSession{}, Attempt{}, err
	}
	result, attempt, err := s.repository.SubmitRushAnswer(
		ctx, userID, sessionID, answer.PuzzleID, answer.Moves, correct,
	)
	if err != nil {
		return RushSession{}, Attempt{}, err
	}
	if err := s.recordPuzzleAchievement(ctx, userID, attempt, correct); err != nil {
		return result, attempt, err
	}
	if result.CurrentPuzzle != nil {
		if err := validateStoredPuzzle(*result.CurrentPuzzle); err != nil {
			return RushSession{}, Attempt{}, err
		}
	}
	return result, attempt, nil
}

func (s *Service) recordPuzzleAchievement(ctx context.Context, userID string, attempt Attempt, correct bool) error {
	if s.achievements == nil || !correct {
		return nil
	}
	if _, err := s.achievements.HandleEvent(ctx, achievements.Event{
		UserID: userID, Type: "puzzle_solved", SourceID: "puzzle-attempt:" + attempt.ID, Amount: 1,
	}); err != nil {
		return fmt.Errorf("record puzzle achievement: %w", err)
	}
	return nil
}

func validateStoredPuzzle(puzzle Puzzle) error {
	position, err := chess.ParseFEN(puzzle.FEN)
	if err != nil {
		return fmt.Errorf("%w: invalid FEN: %v", ErrInvalidPuzzle, err)
	}
	if len(puzzle.Solution) == 0 {
		return fmt.Errorf("%w: solution is empty", ErrInvalidPuzzle)
	}
	for _, text := range puzzle.Solution {
		move, err := chess.ParseMove(text)
		if err != nil {
			return fmt.Errorf("%w: invalid solution move: %v", ErrInvalidPuzzle, err)
		}
		position, err = position.MakeMove(move)
		if err != nil {
			return fmt.Errorf("%w: illegal solution move %q: %v", ErrInvalidPuzzle, text, err)
		}
	}
	return nil
}

func solutionMatches(puzzle Puzzle, submitted []string) (bool, error) {
	position, err := chess.ParseFEN(puzzle.FEN)
	if err != nil {
		return false, fmt.Errorf("%w: invalid FEN: %v", ErrInvalidPuzzle, err)
	}
	matches := len(submitted) == len(puzzle.Solution)
	for index, text := range submitted {
		move, parseErr := chess.ParseMove(text)
		if parseErr != nil {
			return false, nil
		}
		if index >= len(puzzle.Solution) {
			matches = false
		} else if expected, parseExpectedErr := chess.ParseMove(puzzle.Solution[index]); parseExpectedErr != nil {
			return false, fmt.Errorf("%w: invalid stored solution move: %v", ErrInvalidPuzzle, parseExpectedErr)
		} else if move != expected {
			matches = false
		}
		position, err = position.MakeMove(move)
		if err != nil {
			return false, nil
		}
	}
	return matches, nil
}

func validID(id string) bool {
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
