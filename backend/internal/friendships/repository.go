package friendships

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func pair(first, second string) (string, string) {
	first = strings.ToLower(first)
	second = strings.ToLower(second)
	if first < second {
		return first, second
	}
	return second, first
}

func (r *PostgresRepository) List(ctx context.Context, userID string) ([]Friendship, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT CASE WHEN f.user_id = $1 THEN f.friend_id ELSE f.user_id END::text,
			COALESCE(p.username, ''), f.status, f.requested_by::text, f.created_at
		FROM friendships f
		JOIN users other_user ON other_user.id = CASE WHEN f.user_id = $1 THEN f.friend_id ELSE f.user_id END
		LEFT JOIN profiles p ON p.user_id = other_user.id
		WHERE (f.user_id = $1 OR f.friend_id = $1) AND f.status = 'accepted'
		ORDER BY p.username, other_user.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list friendships: %w", err)
	}
	defer rows.Close()
	result := make([]Friendship, 0)
	for rows.Next() {
		var friendship Friendship
		if err := rows.Scan(&friendship.UserID, &friendship.Username, &friendship.Status,
			&friendship.RequestedBy, &friendship.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan friendship: %w", err)
		}
		result = append(result, friendship)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate friendships: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Requests(ctx context.Context, userID string) ([]Friendship, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT f.requested_by::text, COALESCE(p.username, ''), f.status, f.requested_by::text, f.created_at
		FROM friendships f LEFT JOIN profiles p ON p.user_id = f.requested_by
		WHERE (f.user_id = $1 OR f.friend_id = $1)
		  AND f.requested_by <> $1 AND f.status = 'pending'
		ORDER BY f.created_at`, userID)
	if err != nil {
		return nil, fmt.Errorf("list friend requests: %w", err)
	}
	defer rows.Close()
	result := make([]Friendship, 0)
	for rows.Next() {
		var friendship Friendship
		if err := rows.Scan(&friendship.UserID, &friendship.Username, &friendship.Status,
			&friendship.RequestedBy, &friendship.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan friend request: %w", err)
		}
		result = append(result, friendship)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate friend requests: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Request(ctx context.Context, userID, targetID string) error {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, targetID).Scan(&exists); err != nil {
		return fmt.Errorf("check friend request target: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	a, b := pair(userID, targetID)
	result, err := r.db.ExecContext(ctx, `
		INSERT INTO friendships (user_id, friend_id, status, requested_by)
		VALUES ($1, $2, 'pending', $3)
		ON CONFLICT (user_id, friend_id) DO NOTHING`, a, b, userID)
	if err != nil {
		return fmt.Errorf("create friendship request: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read friendship request result: %w", err)
	}
	if affected == 0 {
		return ErrConflict
	}
	return nil
}

func (r *PostgresRepository) Accept(ctx context.Context, userID, requesterID string) error {
	a, b := pair(userID, requesterID)
	result, err := r.db.ExecContext(ctx, `
		UPDATE friendships SET status = 'accepted', updated_at = now()
		WHERE user_id = $1 AND friend_id = $2 AND status = 'pending'
		  AND requested_by = $3`, a, b, requesterID)
	if err != nil {
		return fmt.Errorf("accept friendship: %w", err)
	}
	return oneRowOrConflict(result)
}

func (r *PostgresRepository) Reject(ctx context.Context, userID, requesterID string) error {
	a, b := pair(userID, requesterID)
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM friendships WHERE user_id = $1 AND friend_id = $2
		  AND status = 'pending' AND requested_by = $3`, a, b, requesterID)
	if err != nil {
		return fmt.Errorf("reject friendship: %w", err)
	}
	return oneRowOrConflict(result)
}

func (r *PostgresRepository) Block(ctx context.Context, userID, targetID string) error {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, targetID).Scan(&exists); err != nil {
		return fmt.Errorf("check block target: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	a, b := pair(userID, targetID)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO friendships (user_id, friend_id, status, requested_by)
		VALUES ($1, $2, 'blocked', $3)
		ON CONFLICT (user_id, friend_id) DO UPDATE
		SET status = 'blocked', requested_by = EXCLUDED.requested_by, updated_at = now()`, a, b, userID)
	if err != nil {
		return fmt.Errorf("block user: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Unblock(ctx context.Context, userID, targetID string) error {
	a, b := pair(userID, targetID)
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM friendships WHERE user_id = $1 AND friend_id = $2
		  AND status = 'blocked' AND requested_by = $3`, a, b, userID)
	if err != nil {
		return fmt.Errorf("unblock user: %w", err)
	}
	return oneRowOrConflict(result)
}

func (r *PostgresRepository) DirectAllowed(ctx context.Context, userID, targetID string) error {
	a, b := pair(userID, targetID)
	var allowed bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM friendships
			WHERE user_id = $1 AND friend_id = $2 AND status = 'accepted')`, a, b).Scan(&allowed)
	if err != nil {
		return fmt.Errorf("check direct message permission: %w", err)
	}
	if !allowed {
		return ErrConflict
	}
	return nil
}

func oneRowOrConflict(result sql.Result) error {
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read friendship change result: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

var _ Repository = (*PostgresRepository)(nil)
