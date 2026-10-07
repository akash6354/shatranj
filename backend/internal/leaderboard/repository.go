package leaderboard

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/akash6354/shatranj/backend/internal/ratings"
)

type Repository interface {
	Top(context.Context, ratings.Mode, int) ([]Entry, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Top(ctx context.Context, mode ratings.Mode, limit int) ([]Entry, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.user_id::text, p.username, p.display_name, COALESCE(p.avatar_url, ''),
		       r.mode, r.rating, r.highest, r.games_played, r.wins, r.draws, r.losses, r.updated_at
		FROM ratings r
		JOIN profiles p ON p.user_id = r.user_id
		JOIN users u ON u.id = r.user_id
		WHERE r.mode = $1 AND u.status = 'active'
		ORDER BY r.rating DESC, r.highest DESC, r.updated_at ASC, r.user_id ASC
		LIMIT $2`, mode, limit)
	if err != nil {
		return nil, fmt.Errorf("query leaderboard: %w", err)
	}
	defer rows.Close()

	entries := make([]Entry, 0)
	rank := 1
	for rows.Next() {
		var entry Entry
		if err := rows.Scan(
			&entry.UserID, &entry.Username, &entry.DisplayName, &entry.AvatarURL,
			&entry.Mode, &entry.Rating, &entry.Highest, &entry.Games,
			&entry.Wins, &entry.Draws, &entry.Losses, &entry.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan leaderboard entry: %w", err)
		}
		entry.Rank = rank
		rank++
		entries = append(entries, entry)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate leaderboard: %w", err)
	}
	return entries, nil
}

var _ Repository = (*SQLRepository)(nil)
