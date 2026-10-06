package ratings

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Get(ctx context.Context, userID string, mode Mode) (Record, error) {
	record, err := scanRating(r.db.QueryRowContext(ctx, `
		INSERT INTO ratings (user_id, mode, rating, highest) VALUES ($1, $2, CASE WHEN $2 = 'puzzle' THEN 1500 ELSE 1200 END, CASE WHEN $2 = 'puzzle' THEN 1500 ELSE 1200 END)
		ON CONFLICT (user_id, mode) DO UPDATE SET user_id = EXCLUDED.user_id
		RETURNING user_id::text, mode, rating, highest, games_played, wins, draws, losses, updated_at`,
		userID, mode))
	if err != nil {
		return Record{}, fmt.Errorf("get rating: %w", err)
	}
	return record, nil
}

func (r *SQLRepository) RecordGame(ctx context.Context, game GameResult) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin rating transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.ExecContext(ctx, `
		INSERT INTO rated_game_results (game_id, mode, result, white_user_id, black_user_id)
		VALUES ($1, $2, $3, $4, $5) ON CONFLICT (game_id) DO NOTHING`,
		game.GameID, game.Mode, game.Result, game.WhiteUserID, game.BlackUserID)
	if err != nil {
		return fmt.Errorf("record rated game: %w", err)
	}
	inserted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rated game insert: %w", err)
	}
	if inserted == 0 {
		return tx.Commit()
	}
	userIDs := []string{game.WhiteUserID, game.BlackUserID}
	if userIDs[0] > userIDs[1] {
		userIDs[0], userIDs[1] = userIDs[1], userIDs[0]
	}
	for _, userID := range userIDs {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO ratings (user_id, mode, rating, highest)
			VALUES ($1, $2, CASE WHEN $2 = 'puzzle' THEN 1500 ELSE 1200 END, CASE WHEN $2 = 'puzzle' THEN 1500 ELSE 1200 END)
			ON CONFLICT (user_id, mode) DO NOTHING`, userID, game.Mode); err != nil {
			return fmt.Errorf("initialize rating: %w", err)
		}
	}
	ratings := make(map[string]Record, 2)
	for _, userID := range userIDs {
		record, err := scanRating(tx.QueryRowContext(ctx, `
			SELECT user_id::text, mode, rating, highest, games_played, wins, draws, losses, updated_at
			FROM ratings WHERE user_id = $1 AND mode = $2 FOR UPDATE`, userID, game.Mode))
		if err != nil {
			return fmt.Errorf("lock user rating: %w", err)
		}
		ratings[userID] = record
	}
	whiteScore, err := ScoreForResult(game.Result, true)
	if err != nil {
		return err
	}
	whiteRating, blackRating := ratings[game.WhiteUserID], ratings[game.BlackUserID]
	whiteNew, err := CalculateEloWithK(whiteRating.Rating, blackRating.Rating, whiteScore, kFactor(game.Mode, whiteRating.Games))
	if err != nil {
		return err
	}
	blackNew, err := CalculateEloWithK(blackRating.Rating, whiteRating.Rating, 1-whiteScore, kFactor(game.Mode, blackRating.Games))
	if err != nil {
		return err
	}
	changes := []Change{
		{UserID: game.WhiteUserID, Mode: game.Mode, OldRating: whiteRating.Rating, NewRating: whiteNew},
		{UserID: game.BlackUserID, Mode: game.Mode, OldRating: blackRating.Rating, NewRating: blackNew},
	}
	for _, change := range changes {
		var won, drawn, lost bool
		if change.UserID == game.WhiteUserID {
			won, drawn, lost = game.Result == "1-0", game.Result == "1/2-1/2", game.Result == "0-1"
		} else {
			won, drawn, lost = game.Result == "0-1", game.Result == "1/2-1/2", game.Result == "1-0"
		}
		_, err := tx.ExecContext(ctx, `
			UPDATE ratings SET rating = $3, highest = GREATEST(highest, $3),
				games_played = ratings.games_played + 1,
				wins = wins + $4, draws = draws + $5, losses = losses + $6, updated_at = now()
			WHERE user_id = $1 AND mode = $2`,
			change.UserID, change.Mode, change.NewRating, boolInt(won), boolInt(drawn), boolInt(lost))
		if err != nil {
			return fmt.Errorf("update user rating: %w", err)
		}

	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit rating transaction: %w", err)
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func scanRating(row interface{ Scan(...any) error }) (Record, error) {
	var record Record
	err := row.Scan(&record.UserID, &record.Mode, &record.Rating, &record.Highest,
		&record.Games, &record.Wins, &record.Draws, &record.Losses, &record.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Record{}, fmt.Errorf("rating row not found")
		}
		return Record{}, err
	}
	return record, nil
}
