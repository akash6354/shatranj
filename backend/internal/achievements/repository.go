package achievements

import (
	"context"
	"database/sql"
	"fmt"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) Catalog(ctx context.Context) ([]Definition, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT key, name, description, event_type, threshold, created_at
		FROM achievement_definitions ORDER BY threshold, key`)
	if err != nil {
		return nil, fmt.Errorf("list achievement definitions: %w", err)
	}
	defer rows.Close()
	items := make([]Definition, 0)
	for rows.Next() {
		var item Definition
		if err := rows.Scan(&item.Key, &item.Name, &item.Description, &item.EventType, &item.Threshold, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan achievement definition: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate achievement definitions: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) List(ctx context.Context, userID string) ([]Achievement, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT d.key, d.name, d.description, d.event_type, d.threshold, d.created_at,
			ua.unlocked_at, COALESCE(p.progress, 0)
		FROM achievement_definitions d
		LEFT JOIN user_achievements ua ON ua.achievement_key = d.key AND ua.user_id = $1
		LEFT JOIN user_achievement_progress p ON p.user_id = $1 AND p.event_type = d.event_type
		ORDER BY ua.unlocked_at DESC NULLS LAST, d.threshold, d.key`, userID)
	if err != nil {
		return nil, fmt.Errorf("list user achievements: %w", err)
	}
	defer rows.Close()
	items := make([]Achievement, 0)
	for rows.Next() {
		var item Achievement
		if err := rows.Scan(&item.Key, &item.Name, &item.Description, &item.EventType,
			&item.Threshold, &item.CreatedAt, &item.UnlockedAt, &item.Progress); err != nil {
			return nil, fmt.Errorf("scan user achievement: %w", err)
		}
		item.Unlocked = item.UnlockedAt != nil
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate user achievements: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) ProcessEvent(ctx context.Context, event Event) ([]Achievement, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin achievement event: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO achievement_events (user_id, event_type, source_id)
		VALUES ($1, $2, $3) ON CONFLICT (user_id, event_type, source_id) DO NOTHING`,
		event.UserID, event.Type, event.SourceID)
	if err != nil {
		return nil, fmt.Errorf("record achievement source event: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("read achievement source event result: %w", err)
	}
	if inserted == 0 {
		if err := tx.Commit(); err != nil {
			return nil, fmt.Errorf("commit duplicate achievement event: %w", err)
		}
		return []Achievement{}, nil
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO user_achievement_progress (user_id, event_type, progress, updated_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (user_id, event_type) DO UPDATE
		SET progress = user_achievement_progress.progress + EXCLUDED.progress, updated_at = now()`,
		event.UserID, event.Type, event.Amount); err != nil {
		return nil, fmt.Errorf("record achievement progress: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		WITH eligible AS (
			SELECT d.key, d.name, d.description, d.event_type, d.threshold, d.created_at, p.progress
			FROM achievement_definitions d
			JOIN user_achievement_progress p ON p.event_type = d.event_type
			WHERE p.user_id = $1 AND d.event_type = $2 AND p.progress >= d.threshold
		), inserted AS (
			INSERT INTO user_achievements (user_id, achievement_key)
			SELECT $1, key FROM eligible
			ON CONFLICT (user_id, achievement_key) DO NOTHING
			RETURNING achievement_key, unlocked_at
		)
		SELECT e.key, e.name, e.description, e.event_type, e.threshold, e.created_at,
			i.unlocked_at, e.progress
		FROM inserted i JOIN eligible e ON e.key = i.achievement_key ORDER BY e.key`,
		event.UserID, event.Type)
	if err != nil {
		return nil, fmt.Errorf("unlock achievements: %w", err)
	}
	unlocked := make([]Achievement, 0)
	for rows.Next() {
		var item Achievement
		if err := rows.Scan(&item.Key, &item.Name, &item.Description, &item.EventType,
			&item.Threshold, &item.CreatedAt, &item.UnlockedAt, &item.Progress); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan unlocked achievement: %w", err)
		}
		item.Unlocked = true
		unlocked = append(unlocked, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate unlocked achievements: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close unlocked achievements: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit achievement event: %w", err)
	}
	return unlocked, nil
}
