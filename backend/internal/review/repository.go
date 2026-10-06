package review

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/shatranj/backend/internal/analysis"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) Request(ctx context.Context, gameID, requestedBy string, available bool) (Review, error) {
	initialStatus := StatusUnavailable
	if available {
		initialStatus = StatusPending
	}
	review, err := scanReview(r.db.QueryRowContext(ctx, `
		INSERT INTO game_reviews (game_id, requested_by, status)
		VALUES ($1, $2, $3)
		ON CONFLICT (game_id) DO UPDATE SET
			status = CASE
				WHEN game_reviews.status IN ('failed', 'unavailable') AND $4 THEN 'pending'
				ELSE game_reviews.status
			END,
			requested_by = CASE
				WHEN game_reviews.status IN ('failed', 'unavailable') AND $4 THEN EXCLUDED.requested_by
				ELSE game_reviews.requested_by
			END,
			attempts = CASE
				WHEN game_reviews.status IN ('failed', 'unavailable') AND $4 THEN 0
				ELSE game_reviews.attempts
			END,
			available_at = CASE
				WHEN game_reviews.status IN ('failed', 'unavailable') AND $4 THEN now()
				ELSE game_reviews.available_at
			END,
			error_message = CASE
				WHEN game_reviews.status IN ('failed', 'unavailable') AND $4 THEN NULL
				ELSE game_reviews.error_message
			END,
			updated_at = CASE
				WHEN game_reviews.status IN ('failed', 'unavailable') AND $4 THEN now()
				ELSE game_reviews.updated_at
			END
		RETURNING id::text, game_id::text, status, result, created_at, updated_at`,
		gameID, requestedBy, initialStatus, available))
	if err != nil {
		return Review{}, fmt.Errorf("request game review: %w", err)
	}
	return review, nil
}

func (r *PostgresRepository) Get(ctx context.Context, gameID string) (Review, error) {
	review, err := scanReview(r.db.QueryRowContext(ctx, `
		SELECT id::text, game_id::text, status, result, created_at, updated_at
		FROM game_reviews WHERE game_id = $1`, gameID))
	if errors.Is(err, sql.ErrNoRows) {
		return Review{}, ErrNotFound
	}
	if err != nil {
		return Review{}, fmt.Errorf("get game review: %w", err)
	}
	return review, nil
}

func (r *PostgresRepository) Claim(ctx context.Context) (analysis.Task, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return analysis.Task{}, fmt.Errorf("begin analysis claim: %w", err)
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		UPDATE game_reviews SET status = 'failed', error_message = 'analysis worker stopped unexpectedly',
			updated_at = now()
		WHERE status = 'running' AND attempts >= 3 AND updated_at < now() - interval '2 minutes'`); err != nil {
		return analysis.Task{}, fmt.Errorf("expire abandoned analysis jobs: %w", err)
	}
	var reviewID, gameID string
	err = tx.QueryRowContext(ctx, `
		SELECT id::text, game_id::text
		FROM game_reviews
		WHERE (status = 'pending' AND available_at <= now())
			OR (status = 'running' AND attempts < 3 AND updated_at < now() - interval '2 minutes')
		ORDER BY created_at, id
		LIMIT 1
		FOR UPDATE SKIP LOCKED`).Scan(&reviewID, &gameID)
	if errors.Is(err, sql.ErrNoRows) {
		return analysis.Task{}, analysis.ErrNoJob
	}
	if err != nil {
		return analysis.Task{}, fmt.Errorf("select analysis job: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE game_reviews SET status = 'running', attempts = attempts + 1,
			error_message = NULL, updated_at = now()
		WHERE id = $1`, reviewID); err != nil {
		return analysis.Task{}, fmt.Errorf("mark analysis job running: %w", err)
	}
	var task analysis.Task
	task.ReviewID, task.GameID = reviewID, gameID
	err = tx.QueryRowContext(ctx, `
		SELECT white_player_id::text, COALESCE(black_player_id::text, '')
		FROM games WHERE id = $1 AND status = 'finished'`, gameID).Scan(&task.WhiteID, &task.BlackID)
	if errors.Is(err, sql.ErrNoRows) {
		return analysis.Task{}, fmt.Errorf("review game is not finished")
	}
	if err != nil {
		return analysis.Task{}, fmt.Errorf("load players for review: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT player_id::text, uci, san, fen_after
		FROM game_moves WHERE game_id = $1 ORDER BY ply`, gameID)
	if err != nil {
		return analysis.Task{}, fmt.Errorf("load review game moves: %w", err)
	}
	task.Moves = make([]analysis.PlayedMove, 0)
	for rows.Next() {
		var move analysis.PlayedMove
		if err := rows.Scan(&move.PlayerID, &move.UCI, &move.SAN, &move.FENAfter); err != nil {
			rows.Close()
			return analysis.Task{}, fmt.Errorf("scan review game move: %w", err)
		}
		task.Moves = append(task.Moves, move)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return analysis.Task{}, fmt.Errorf("iterate review game moves: %w", err)
	}
	if err := rows.Close(); err != nil {
		return analysis.Task{}, fmt.Errorf("close review game moves: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return analysis.Task{}, fmt.Errorf("commit analysis claim: %w", err)
	}
	return task, nil
}

func (r *PostgresRepository) Complete(ctx context.Context, result analysis.Result) error {
	encoded, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("encode analysis result: %w", err)
	}
	update, err := r.db.ExecContext(ctx, `
		UPDATE game_reviews
		SET status = 'completed', result = $2::jsonb, error_message = NULL, updated_at = now()
		WHERE game_id = $1 AND status = 'running'`, result.GameID, string(encoded))
	if err != nil {
		return fmt.Errorf("complete game review: %w", err)
	}
	if err := requireOneReviewRow(update); err != nil {
		return fmt.Errorf("complete game review: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Fail(ctx context.Context, reviewID string, cause error) error {
	message := strings.TrimSpace(cause.Error())
	if len(message) > 1000 {
		message = message[:1000]
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE game_reviews
		SET status = CASE WHEN attempts >= 3 THEN 'failed' ELSE 'pending' END,
			error_message = $2,
			available_at = CASE WHEN attempts >= 3 THEN available_at ELSE now() + interval '5 seconds' END,
			updated_at = now()
		WHERE id = $1 AND status = 'running'`, reviewID, message)
	if err != nil {
		return fmt.Errorf("record game review failure: %w", err)
	}
	return nil
}

func (r *PostgresRepository) MarkUnavailable(ctx context.Context, reviewID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE game_reviews SET status = 'unavailable', error_message = 'Stockfish is not configured',
			updated_at = now()
		WHERE id = $1 AND status = 'running'`, reviewID)
	if err != nil {
		return fmt.Errorf("mark game review unavailable: %w", err)
	}
	return nil
}

func scanReview(row interface{ Scan(...any) error }) (Review, error) {
	var review Review
	var status string
	var encoded []byte
	err := row.Scan(&review.ID, &review.GameID, &status, &encoded, &review.CreatedAt, &review.UpdatedAt)
	if err != nil {
		return Review{}, err
	}
	review.Status = Status(status)
	if len(encoded) > 0 {
		var result analysis.Result
		if err := json.Unmarshal(encoded, &result); err != nil {
			return Review{}, fmt.Errorf("decode game review result: %w", err)
		}
		review.Result = &result
	}
	return review, nil
}

func requireOneReviewRow(result sql.Result) error {
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected != 1 {
		return ErrNotFound
	}
	return nil
}

var _ Repository = (*PostgresRepository)(nil)
var _ JobRepository = (*PostgresRepository)(nil)
