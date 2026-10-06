package puzzles

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) List(ctx context.Context, filter Filter) ([]Puzzle, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT `+puzzleColumns+`
		FROM puzzles
		WHERE status = 'published'
		  AND ($1 = '' OR themes ? $1)
		  AND ($2 = '' OR difficulty = $2)
		ORDER BY rating, id
		LIMIT $3 OFFSET $4`, filter.Theme, filter.Difficulty, filter.Limit, filter.Offset)
	if err != nil {
		return nil, fmt.Errorf("list puzzles: %w", err)
	}
	defer rows.Close()
	puzzles := make([]Puzzle, 0)
	for rows.Next() {
		puzzle, err := scanPuzzle(rows)
		if err != nil {
			return nil, fmt.Errorf("scan puzzle: %w", err)
		}
		puzzles = append(puzzles, puzzle)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate puzzles: %w", err)
	}
	return puzzles, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Puzzle, error) {
	puzzle, err := scanPuzzle(r.db.QueryRowContext(ctx, `
		SELECT `+puzzleColumns+` FROM puzzles WHERE id = $1 AND status = 'published'`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Puzzle{}, ErrNotFound
	}
	if err != nil {
		return Puzzle{}, fmt.Errorf("get puzzle: %w", err)
	}
	return puzzle, nil
}

func (r *PostgresRepository) Daily(ctx context.Context, date time.Time) (Puzzle, error) {
	date = date.UTC()
	puzzle, err := scanPuzzle(r.db.QueryRowContext(ctx, `
		SELECT `+puzzleColumns+`
		FROM puzzles
		WHERE id = COALESCE(
			(SELECT puzzle_id FROM daily_puzzles WHERE puzzle_date = $1::date),
			(SELECT id FROM puzzles WHERE status = 'published' ORDER BY md5(id::text || $1::date::text) LIMIT 1)
		) AND status = 'published'`, dailyDate(date)))
	if errors.Is(err, sql.ErrNoRows) {
		return Puzzle{}, ErrNotFound
	}
	if err != nil {
		return Puzzle{}, fmt.Errorf("get daily puzzle: %w", err)
	}
	return puzzle, nil
}

func (r *PostgresRepository) Rating(ctx context.Context, userID string) (int, error) {
	var rating int
	err := r.db.QueryRowContext(ctx, `
		SELECT rating FROM puzzle_ratings WHERE user_id = $1`, userID).Scan(&rating)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultPuzzleRating, nil
	}
	if err != nil {
		return 0, fmt.Errorf("get puzzle rating: %w", err)
	}
	return rating, nil
}

func (r *PostgresRepository) RecordAttempt(ctx context.Context, input AttemptInput) (Attempt, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Attempt{}, fmt.Errorf("begin puzzle attempt: %w", err)
	}
	defer tx.Rollback()
	attempt, err := recordAttemptTx(ctx, tx, input)
	if err != nil {
		return Attempt{}, err
	}
	if err := tx.Commit(); err != nil {
		return Attempt{}, fmt.Errorf("commit puzzle attempt: %w", err)
	}
	return attempt, nil
}

func (r *PostgresRepository) StartRush(ctx context.Context, userID string, durationSeconds int) (RushSession, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RushSession{}, fmt.Errorf("begin puzzle rush: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO puzzle_ratings (user_id, rating) VALUES ($1, $2)
		ON CONFLICT (user_id) DO NOTHING`, userID, defaultPuzzleRating); err != nil {
		return RushSession{}, fmt.Errorf("initialize puzzle rating: %w", err)
	}
	var puzzleID string
	err = tx.QueryRowContext(ctx, `
		SELECT p.id
		FROM puzzles p
		JOIN puzzle_ratings r ON r.user_id = $1
		WHERE p.status = 'published'
		ORDER BY abs(p.rating - r.rating), md5(p.id::text || now()::text)
		LIMIT 1`, userID).Scan(&puzzleID)
	if errors.Is(err, sql.ErrNoRows) {
		return RushSession{}, ErrNotFound
	}
	if err != nil {
		return RushSession{}, fmt.Errorf("select rush puzzle: %w", err)
	}
	session, err := scanRush(tx.QueryRowContext(ctx, `
		INSERT INTO puzzle_rush_sessions
			(user_id, current_puzzle_id, time_limit_seconds, started_at, expires_at)
		VALUES ($1, $2, $3, now(), now() + ($3 * interval '1 second'))
		RETURNING id, user_id, status, score, streak, time_limit_seconds,
			current_puzzle_id, started_at, expires_at, finished_at`,
		userID, puzzleID, durationSeconds))
	if err != nil {
		return RushSession{}, fmt.Errorf("create rush session: %w", err)
	}
	initialPuzzle, err := scanPuzzle(tx.QueryRowContext(ctx, `
		SELECT `+puzzleColumns+` FROM puzzles WHERE id = $1`, puzzleID))
	if err != nil {
		return RushSession{}, fmt.Errorf("load initial rush puzzle: %w", err)
	}
	session.CurrentPuzzle = &initialPuzzle
	if err := tx.Commit(); err != nil {
		return RushSession{}, fmt.Errorf("commit puzzle rush: %w", err)
	}
	return session, nil
}

func (r *PostgresRepository) GetRush(ctx context.Context, userID, sessionID string) (RushSession, error) {
	session, err := scanRush(r.db.QueryRowContext(ctx, `
		SELECT id, user_id, status, score, streak, time_limit_seconds,
			current_puzzle_id, started_at, expires_at, finished_at
		FROM puzzle_rush_sessions WHERE id = $1 AND user_id = $2`, sessionID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return RushSession{}, ErrRushNotFound
	}
	if err != nil {
		return RushSession{}, fmt.Errorf("get puzzle rush: %w", err)
	}
	if session.CurrentPuzzleID != "" {
		puzzle, err := scanPuzzle(r.db.QueryRowContext(ctx, `
			SELECT `+puzzleColumns+` FROM puzzles WHERE id = $1`, session.CurrentPuzzleID))
		if err != nil {
			return RushSession{}, fmt.Errorf("load current rush puzzle: %w", err)
		}
		session.CurrentPuzzle = &puzzle
	}
	if session.Status == RushActive {
		var finishedAt sql.NullTime
		err := r.db.QueryRowContext(ctx, `
			UPDATE puzzle_rush_sessions SET status = 'expired', finished_at = now()
			WHERE id = $1 AND status = 'active' AND expires_at <= now()
			RETURNING finished_at`, session.ID).Scan(&finishedAt)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return RushSession{}, fmt.Errorf("expire puzzle rush: %w", err)
		}
		if finishedAt.Valid {
			session.Status = RushExpired
			session.FinishedAt = &finishedAt.Time
		}
	}
	return session, nil
}

func (r *PostgresRepository) SubmitRushAnswer(
	ctx context.Context, userID, sessionID, puzzleID string, moves []string, correct bool,
) (RushSession, Attempt, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return RushSession{}, Attempt{}, fmt.Errorf("begin rush answer: %w", err)
	}
	defer tx.Rollback()
	session, err := scanRush(tx.QueryRowContext(ctx, `
		SELECT id, user_id, status, score, streak, time_limit_seconds,
			current_puzzle_id, started_at, expires_at, finished_at
		FROM puzzle_rush_sessions WHERE id = $1 AND user_id = $2 FOR UPDATE`, sessionID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return RushSession{}, Attempt{}, ErrRushNotFound
	}
	if err != nil {
		return RushSession{}, Attempt{}, fmt.Errorf("lock rush session: %w", err)
	}
	var databaseNow time.Time
	if err := tx.QueryRowContext(ctx, `SELECT now()`).Scan(&databaseNow); err != nil {
		return RushSession{}, Attempt{}, fmt.Errorf("check rush session time: %w", err)
	}
	if session.Status != RushActive || !databaseNow.Before(session.ExpiresAt) {
		if session.Status == RushActive {
			_, err = tx.ExecContext(ctx, `UPDATE puzzle_rush_sessions SET status = 'expired', finished_at = now() WHERE id = $1`, sessionID)
			if err != nil {
				return RushSession{}, Attempt{}, fmt.Errorf("expire rush session: %w", err)
			}
			session.Status = RushExpired
			if err := tx.Commit(); err != nil {
				return RushSession{}, Attempt{}, fmt.Errorf("commit expired rush: %w", err)
			}
		}
		return RushSession{}, Attempt{}, ErrRushFinished
	}
	if session.CurrentPuzzleID != puzzleID {
		return RushSession{}, Attempt{}, ErrRushPuzzle
	}
	attempt, err := recordAttemptTx(ctx, tx, AttemptInput{
		UserID: userID, PuzzleID: puzzleID, RushSessionID: sessionID, Moves: moves, Correct: correct,
	})
	if err != nil {
		return RushSession{}, Attempt{}, err
	}

	var nextPuzzleID string
	if correct {
		session.Score++
		session.Streak++
		err = tx.QueryRowContext(ctx, `
			SELECT p.id
			FROM puzzles p
			JOIN puzzle_ratings r ON r.user_id = $1
			WHERE p.status = 'published' AND p.id <> $2
			ORDER BY abs(p.rating - r.rating), md5(p.id::text || now()::text)
			LIMIT 1`, userID, puzzleID).Scan(&nextPuzzleID)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			return RushSession{}, Attempt{}, fmt.Errorf("select next rush puzzle: %w", err)
		}
		if nextPuzzleID == "" {
			session.Status = RushFinished
			session.FinishedAt = timePointer(databaseNow)
		} else {
			session.CurrentPuzzleID = nextPuzzleID
		}
	} else {
		session.Streak = 0
		session.Status = RushFinished
		session.FinishedAt = timePointer(databaseNow)
	}
	_, err = tx.ExecContext(ctx, `
		UPDATE puzzle_rush_sessions
		SET score = $2, streak = $3, status = $4, current_puzzle_id = NULLIF($5, '')::uuid,
			finished_at = $6
		WHERE id = $1`, sessionID, session.Score, session.Streak, session.Status,
		session.CurrentPuzzleID, session.FinishedAt)
	if err != nil {
		return RushSession{}, Attempt{}, fmt.Errorf("update rush session: %w", err)
	}
	if session.CurrentPuzzleID != "" {
		puzzle, err := scanPuzzle(tx.QueryRowContext(ctx, `
			SELECT `+puzzleColumns+` FROM puzzles WHERE id = $1`, session.CurrentPuzzleID))
		if err != nil {
			return RushSession{}, Attempt{}, fmt.Errorf("load next rush puzzle: %w", err)
		}
		session.CurrentPuzzle = &puzzle
	}
	if err := tx.Commit(); err != nil {
		return RushSession{}, Attempt{}, fmt.Errorf("commit rush answer: %w", err)
	}
	return session, attempt, nil
}

const puzzleColumns = `id, fen, solution_moves, themes, difficulty, rating,
	explanation, status, created_at, updated_at`

type scanner interface {
	Scan(...any) error
}

func scanPuzzle(row scanner) (Puzzle, error) {
	var puzzle Puzzle
	var solutionJSON, themesJSON []byte
	err := row.Scan(&puzzle.ID, &puzzle.FEN, &solutionJSON, &themesJSON, &puzzle.Difficulty,
		&puzzle.Rating, &puzzle.Explanation, &puzzle.Status, &puzzle.CreatedAt, &puzzle.UpdatedAt)
	if err != nil {
		return Puzzle{}, err
	}
	if err := json.Unmarshal(solutionJSON, &puzzle.Solution); err != nil {
		return Puzzle{}, fmt.Errorf("decode puzzle solution: %w", err)
	}
	if err := json.Unmarshal(themesJSON, &puzzle.Themes); err != nil {
		return Puzzle{}, fmt.Errorf("decode puzzle themes: %w", err)
	}
	return puzzle, nil
}

func scanRush(row scanner) (RushSession, error) {
	var session RushSession
	var finishedAt sql.NullTime
	err := row.Scan(&session.ID, &session.UserID, &session.Status, &session.Score, &session.Streak,
		&session.TimeLimitSeconds, &session.CurrentPuzzleID, &session.StartedAt, &session.ExpiresAt, &finishedAt)
	if err != nil {
		return RushSession{}, err
	}
	if finishedAt.Valid {
		session.FinishedAt = &finishedAt.Time
	}
	return session, nil
}

func timePointer(value time.Time) *time.Time {
	return &value
}

func recordAttemptTx(ctx context.Context, tx *sql.Tx, input AttemptInput) (Attempt, error) {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO puzzle_ratings (user_id, rating) VALUES ($1, $2)
		ON CONFLICT (user_id) DO NOTHING`, input.UserID, defaultPuzzleRating); err != nil {
		return Attempt{}, fmt.Errorf("initialize puzzle rating: %w", err)
	}
	var userRating, puzzleRating int
	if err := tx.QueryRowContext(ctx, `
		SELECT rating FROM puzzle_ratings WHERE user_id = $1 FOR UPDATE`, input.UserID).Scan(&userRating); err != nil {
		return Attempt{}, fmt.Errorf("lock user puzzle rating: %w", err)
	}
	if err := tx.QueryRowContext(ctx, `
		SELECT rating FROM puzzles WHERE id = $1 AND status = 'published' FOR UPDATE`, input.PuzzleID).Scan(&puzzleRating); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Attempt{}, ErrNotFound
		}
		return Attempt{}, fmt.Errorf("lock puzzle rating: %w", err)
	}
	userAfter, puzzleAfter, err := updatedPuzzleRatings(userRating, puzzleRating, input.Correct)
	if err != nil {
		return Attempt{}, fmt.Errorf("calculate puzzle ratings: %w", err)
	}
	movesJSON, err := json.Marshal(input.Moves)
	if err != nil {
		return Attempt{}, fmt.Errorf("encode puzzle answer: %w", err)
	}
	var attempt Attempt
	err = tx.QueryRowContext(ctx, `
		INSERT INTO puzzle_attempts
			(user_id, puzzle_id, rush_session_id, submitted_moves, correct, user_rating_before, user_rating_after,
			 puzzle_rating_before, puzzle_rating_after)
		VALUES ($1, $2, NULLIF($3, '')::uuid, $4::jsonb, $5, $6, $7, $8, $9)
		RETURNING id, puzzle_id, correct, submitted_moves, user_rating_before, user_rating_after,
			puzzle_rating_before, puzzle_rating_after, created_at`,
		input.UserID, input.PuzzleID, input.RushSessionID, string(movesJSON), input.Correct,
		userRating, userAfter, puzzleRating, puzzleAfter).Scan(
		&attempt.ID, &attempt.PuzzleID, &attempt.Correct, &movesJSON,
		&attempt.RatingBefore, &attempt.RatingAfter, &attempt.PuzzleRatingBefore,
		&attempt.PuzzleRatingAfter, &attempt.CreatedAt)
	if err != nil {
		return Attempt{}, fmt.Errorf("record puzzle attempt: %w", err)
	}
	if err := json.Unmarshal(movesJSON, &attempt.Moves); err != nil {
		return Attempt{}, fmt.Errorf("decode submitted puzzle moves: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE puzzle_ratings SET rating = $2, highest = GREATEST(highest, $2),
			games_played = games_played + 1, correct = correct + $3, updated_at = now()
		WHERE user_id = $1`, input.UserID, userAfter, boolInt(input.Correct)); err != nil {
		return Attempt{}, fmt.Errorf("update user puzzle rating: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE puzzles SET rating = $2, updated_at = now() WHERE id = $1`, input.PuzzleID, puzzleAfter); err != nil {
		return Attempt{}, fmt.Errorf("update puzzle rating: %w", err)
	}
	return attempt, nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
