package notifications

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) Create(ctx context.Context, input CreateInput) (Notification, error) {
	data, err := json.Marshal(input.Data)
	if err != nil {
		return Notification{}, fmt.Errorf("encode notification data: %w", err)
	}
	return scanNotification(r.db.QueryRowContext(ctx, `
		INSERT INTO notifications (user_id, type, title, body, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, user_id::text, type, title, body, data, read_at, created_at`,
		input.UserID, input.Type, input.Title, input.Body, data))
}

func (r *PostgresRepository) List(ctx context.Context, userID string, unreadOnly bool, limit, offset int) ([]Notification, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, user_id::text, type, title, body, data, read_at, created_at
		FROM notifications WHERE user_id = $1 AND (NOT $2 OR read_at IS NULL)
		ORDER BY created_at DESC, id DESC LIMIT $3 OFFSET $4`, userID, unreadOnly, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	defer rows.Close()
	items := make([]Notification, 0)
	for rows.Next() {
		item, err := scanNotification(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate notifications: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) UnreadCount(ctx context.Context, userID string) (int, error) {
	var count int
	if err := r.db.QueryRowContext(ctx, `
		SELECT count(*) FROM notifications WHERE user_id = $1 AND read_at IS NULL`, userID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count unread notifications: %w", err)
	}
	return count, nil
}

func (r *PostgresRepository) SetRead(ctx context.Context, userID, id string, read bool) (Notification, error) {
	var row *sql.Row
	if read {
		row = r.db.QueryRowContext(ctx, `
			UPDATE notifications SET read_at = COALESCE(read_at, now())
			WHERE id = $1 AND user_id = $2
			RETURNING id::text, user_id::text, type, title, body, data, read_at, created_at`, id, userID)
	} else {
		row = r.db.QueryRowContext(ctx, `
			UPDATE notifications SET read_at = NULL WHERE id = $1 AND user_id = $2
			RETURNING id::text, user_id::text, type, title, body, data, read_at, created_at`, id, userID)
	}
	item, err := scanNotification(row)
	if errors.Is(err, sql.ErrNoRows) {
		return Notification{}, ErrNotFound
	}
	return item, err
}

type scanner interface{ Scan(...any) error }

func scanNotification(row scanner) (Notification, error) {
	var item Notification
	var data []byte
	err := row.Scan(&item.ID, &item.UserID, &item.Type, &item.Title, &item.Body, &data, &item.ReadAt, &item.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Notification{}, ErrNotFound
	}
	if err != nil {
		return Notification{}, fmt.Errorf("read notification: %w", err)
	}
	if len(data) != 0 {
		if err := json.Unmarshal(data, &item.Data); err != nil {
			return Notification{}, fmt.Errorf("decode notification data: %w", err)
		}
	}
	return item, nil
}
