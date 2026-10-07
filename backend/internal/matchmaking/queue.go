// Package matchmaking pairs players for new games.
//
// The queue is PostgreSQL-backed, not Redis-backed: JoinQueue inserts an entry
// and finds an opponent inside one transaction guarded by a pg advisory lock,
// so it is durable and safe across processes without a shared cache. The
// in-memory fallback named in the Redis integration plan is not used here
// because the database is already the single source of truth for the queue;
// layering a Redis queue on top would risk the two views diverging. If Redis
// queue/state is added later it must mirror this repository rather than replace
// it.
package matchmaking

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
)

// SQLQueueRepository is the PostgreSQL-backed matchmaking queue.
type SQLQueueRepository struct {
	db  *sql.DB
	now func() time.Time
}

func NewQueueRepository(db *sql.DB) *SQLQueueRepository {
	return &SQLQueueRepository{db: db, now: time.Now}
}

func (r *SQLQueueRepository) JoinQueue(ctx context.Context, entry Entry) (Entry, *Entry, error) {
	id, err := randomUUID()
	if err != nil {
		return Entry{}, nil, err
	}
	entry.ID = id
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Entry{}, nil, fmt.Errorf("begin matchmaking join: %w", err)
	}
	defer tx.Rollback()
	lockKey := fmt.Sprintf("%s:%d:%d", entry.Mode, entry.TimeControl.InitialSeconds, entry.TimeControl.IncrementSeconds)
	if _, err := tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, lockKey); err != nil {
		return Entry{}, nil, fmt.Errorf("lock matchmaking pool: %w", err)
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO matchmaking_queue (
			id, user_id, mode, initial_seconds, increment_seconds, rating
		) VALUES ($1, $2, $3, $4, $5, $6)`,
		entry.ID, entry.UserID, entry.Mode, entry.TimeControl.InitialSeconds,
		entry.TimeControl.IncrementSeconds, entry.Rating)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" {
			return Entry{}, nil, ErrAlreadyQueued
		}
		return Entry{}, nil, fmt.Errorf("insert matchmaking entry: %w", err)
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT id::text, user_id::text, mode, initial_seconds, increment_seconds,
		       rating, status, COALESCE(match_id::text, ''), COALESCE(game_id::text, ''),
		       created_at, updated_at
		FROM matchmaking_queue
		WHERE status = 'queued' AND user_id <> $1 AND mode = $2
		  AND initial_seconds = $3 AND increment_seconds = $4
		ORDER BY created_at
		LIMIT 500
		FOR UPDATE SKIP LOCKED`,
		entry.UserID, entry.Mode, entry.TimeControl.InitialSeconds, entry.TimeControl.IncrementSeconds)
	if err != nil {
		return Entry{}, nil, fmt.Errorf("find matchmaking candidates: %w", err)
	}
	now := r.now()
	var candidate *Entry
	for rows.Next() {
		var possible Entry
		if err := rows.Scan(&possible.ID, &possible.UserID, &possible.Mode,
			&possible.TimeControl.InitialSeconds, &possible.TimeControl.IncrementSeconds,
			&possible.Rating, &possible.Status, &possible.MatchID, &possible.GameID,
			&possible.CreatedAt, &possible.UpdatedAt); err != nil {
			_ = rows.Close()
			return Entry{}, nil, fmt.Errorf("scan matchmaking candidate: %w", err)
		}
		if ratingsCompatible(entry.Rating, 0, possible.Rating, now.Sub(possible.CreatedAt)) {
			candidate = &possible
			break
		}
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return Entry{}, nil, fmt.Errorf("iterate matchmaking candidates: %w", err)
	}
	if err := rows.Close(); err != nil {
		return Entry{}, nil, fmt.Errorf("close matchmaking candidates: %w", err)
	}
	if candidate == nil {
		if err := tx.Commit(); err != nil {
			return Entry{}, nil, fmt.Errorf("commit matchmaking entry: %w", err)
		}
		entry.Status = StatusQueued
		entry.CreatedAt, entry.UpdatedAt = now, now
		entry.RatingRange = RatingRangeForWait(0)
		return entry, nil, nil
	}
	matchID, err := randomUUID()
	if err != nil {
		return Entry{}, nil, err
	}
	reservation, err := tx.ExecContext(ctx, `
		UPDATE matchmaking_queue SET status = 'matched', match_id = $3, updated_at = now()
		WHERE id IN ($1, $2) AND status = 'queued'`, entry.ID, candidate.ID, matchID)
	if err != nil {
		return Entry{}, nil, fmt.Errorf("reserve matchmaking pair: %w", err)
	}
	reserved, err := reservation.RowsAffected()
	if err != nil {
		return Entry{}, nil, fmt.Errorf("check matchmaking reservation: %w", err)
	}
	if reserved != 2 {
		return Entry{}, nil, fmt.Errorf("reserve matchmaking pair: expected 2 entries, updated %d", reserved)
	}
	if err := tx.Commit(); err != nil {
		return Entry{}, nil, fmt.Errorf("commit matchmaking pair: %w", err)
	}
	entry.Status, entry.MatchID, entry.CreatedAt, entry.UpdatedAt = StatusMatched, matchID, now, now
	entry.RatingRange = RatingRangeForWait(0)
	candidate.Status, candidate.MatchID, candidate.UpdatedAt = StatusMatched, matchID, now
	candidate.RatingRange = RatingRangeForWait(now.Sub(candidate.CreatedAt))
	return entry, candidate, nil
}

func (r *SQLQueueRepository) LeaveQueue(ctx context.Context, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE matchmaking_queue
		SET status = 'cancelled', match_id = NULL, game_id = NULL, updated_at = now()
		WHERE user_id = $1 AND (
			status = 'queued'
			OR (status = 'paired' AND EXISTS (
				SELECT 1 FROM games WHERE games.id = matchmaking_queue.game_id AND games.status = 'finished'
			))
		)`, userID)
	if err != nil {
		return fmt.Errorf("leave matchmaking queue: %w", err)
	}
	changed, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check queue leave: %w", err)
	}
	if changed == 0 {
		var status Status
		err := r.db.QueryRowContext(ctx, `
			SELECT status FROM matchmaking_queue
			WHERE user_id = $1 AND status IN ('matched', 'paired')
			ORDER BY created_at DESC LIMIT 1`, userID).Scan(&status)
		if err == nil {
			return ErrMatchInProgress
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("check matchmaking status while leaving: %w", err)
		}
		return ErrQueueNotFound
	}
	return nil
}

func (r *SQLQueueRepository) GetQueueStatus(ctx context.Context, userID string) (Entry, error) {
	var entry Entry
	err := r.db.QueryRowContext(ctx, `
		SELECT id::text, user_id::text, mode, initial_seconds, increment_seconds,
		       rating, status, COALESCE(match_id::text, ''), COALESCE(game_id::text, ''),
		       created_at, updated_at
		FROM matchmaking_queue WHERE user_id = $1 AND status IN ('queued', 'matched', 'paired')
		ORDER BY created_at DESC LIMIT 1`, userID).Scan(&entry.ID, &entry.UserID, &entry.Mode,
		&entry.TimeControl.InitialSeconds, &entry.TimeControl.IncrementSeconds,
		&entry.Rating, &entry.Status, &entry.MatchID, &entry.GameID,
		&entry.CreatedAt, &entry.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Entry{}, ErrQueueNotFound
	}
	if err != nil {
		return Entry{}, fmt.Errorf("get matchmaking status: %w", err)
	}
	entry.RatingRange = RatingRangeForWait(r.now().Sub(entry.CreatedAt))
	return entry, nil
}

func (r *SQLQueueRepository) CompleteMatch(ctx context.Context, matchID, gameID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE matchmaking_queue SET status = 'paired', game_id = $2, updated_at = now()
		WHERE match_id = $1 AND status = 'matched'`, matchID, gameID)
	if err != nil {
		return fmt.Errorf("complete matchmaking pair: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check completed pair: %w", err)
	}
	if count != 2 {
		return fmt.Errorf("complete matchmaking pair: expected 2 entries, updated %d", count)
	}
	return nil
}

func (r *SQLQueueRepository) AbortMatch(ctx context.Context, matchID string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE matchmaking_queue SET status = 'cancelled', match_id = NULL, game_id = NULL, updated_at = now()
		WHERE match_id = $1 AND status = 'matched'`, matchID)
	if err != nil {
		return fmt.Errorf("abort matchmaking pair: %w", err)
	}
	return nil
}

func randomUUID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("generate matchmaking ID: %w", err)
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16]), nil
}
