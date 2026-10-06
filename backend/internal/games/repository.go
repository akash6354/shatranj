package games

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/shatranj/backend/internal/chess"
)

type Repository interface {
	Create(context.Context, string, CreateInput) (Game, error)
	Join(context.Context, string, string) (Game, error)
	Get(context.Context, string) (Game, error)
	SaveMove(context.Context, string, string, string, MoveUpdate) (Game, error)
	Resign(context.Context, string, string) (Game, error)
	OfferDraw(context.Context, string, string) (Game, error)
	AcceptDraw(context.Context, string, string) (Game, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) Create(ctx context.Context, userID string, input CreateInput) (Game, error) {
	id, err := newGameID()
	if err != nil {
		return Game{}, err
	}
	status := StatusWaiting
	game, err := scanGame(r.db.QueryRowContext(ctx, `
		INSERT INTO games (
			id, white_player_id, time_control_initial_seconds, time_control_increment_seconds,
			mode, rated, current_fen, status, result
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '*')
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
			created_at, started_at, finished_at, updated_at`,
		id, userID, input.TimeControl.InitialSeconds, input.TimeControl.IncrementSeconds,
		input.Mode, input.Rated, chess.StartingPosition().FEN(), status))
	if err != nil {
		return Game{}, fmt.Errorf("create game: %w", err)
	}
	game.Moves = []MoveRecord{}
	return game, nil
}

func (r *SQLRepository) Join(ctx context.Context, gameID, userID string) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, fmt.Errorf("begin join game: %w", err)
	}
	defer tx.Rollback()
	game, err := scanGame(tx.QueryRowContext(ctx, gameSelect+` WHERE id = $1 FOR UPDATE`, gameID))
	if err != nil {
		return Game{}, err
	}
	if game.WhitePlayerID == userID {
		return Game{}, ErrCannotJoinOwn
	}
	if game.Status != StatusWaiting || game.BlackPlayerID != "" {
		return Game{}, ErrGameFull
	}
	game, err = scanGame(tx.QueryRowContext(ctx, `
		UPDATE games SET black_player_id = $2, status = 'active',
			started_at = now(), updated_at = now()
		WHERE id = $1 AND status = 'waiting' AND black_player_id IS NULL
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
			created_at, started_at, finished_at, updated_at`, gameID, userID))
	if err != nil {
		return Game{}, fmt.Errorf("join game: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Game{}, fmt.Errorf("commit join game: %w", err)
	}
	game.Moves = []MoveRecord{}
	return game, nil
}

func (r *SQLRepository) Get(ctx context.Context, gameID string) (Game, error) {
	game, err := scanGame(r.db.QueryRowContext(ctx, gameSelect+` WHERE id = $1`, gameID))
	if err != nil {
		return Game{}, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT move_number, player_id::text, uci, san, fen_after,
		       white_clock_ms, black_clock_ms, created_at
		FROM game_moves WHERE game_id = $1 ORDER BY ply`, gameID)
	if err != nil {
		return Game{}, fmt.Errorf("list game moves: %w", err)
	}
	defer rows.Close()
	game.Moves = make([]MoveRecord, 0)
	for rows.Next() {
		var move MoveRecord
		if err := rows.Scan(&move.Number, &move.PlayerID, &move.UCI, &move.SAN, &move.FENAfter,
			&move.WhiteClockMS, &move.BlackClockMS, &move.CreatedAt); err != nil {
			return Game{}, fmt.Errorf("scan game move: %w", err)
		}
		game.Moves = append(game.Moves, move)
	}
	if err := rows.Err(); err != nil {
		return Game{}, fmt.Errorf("iterate game moves: %w", err)
	}
	return game, nil
}

func (r *SQLRepository) SaveMove(ctx context.Context, gameID, userID, expectedFEN string, update MoveUpdate) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, fmt.Errorf("begin move transaction: %w", err)
	}
	defer tx.Rollback()

	var game Game
	var startedAt, finishedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		UPDATE games SET current_fen = $4, status = $5, result = $6, end_reason = NULLIF($7, ''),
			finished_at = CASE WHEN $5 = 'finished' THEN now() ELSE NULL END, updated_at = now()
		WHERE id = $1 AND current_fen = $2 AND status = 'active'
			AND (white_player_id = $3 OR black_player_id = $3)
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
			created_at, started_at, finished_at, updated_at`,
		gameID, expectedFEN, userID, update.FEN, update.Status, update.Result, update.Reason,
	).Scan(&game.ID, &game.WhitePlayerID, &game.BlackPlayerID, &game.TimeControl.InitialSeconds,
		&game.TimeControl.IncrementSeconds, &game.Mode, &game.Rated, &game.CurrentFEN, &game.Status, &game.Result, &game.EndReason,
		&game.DrawOfferBy, &game.CreatedAt, &startedAt, &finishedAt, &game.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrMoveConflict
	}
	if err != nil {
		return Game{}, fmt.Errorf("update game position: %w", err)
	}
	game.StartedAt = nullTimePtr(startedAt)
	game.FinishedAt = nullTimePtr(finishedAt)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO game_moves (
			game_id, ply, move_number, player_id, uci, san, fen_after, white_clock_ms, black_clock_ms
		) VALUES ($1, (SELECT COALESCE(MAX(ply), 0) + 1 FROM game_moves WHERE game_id = $1),
			$2, $3, $4, $5, $6, $7, $8)`,
		gameID, update.Record.Number, userID, update.Record.UCI, update.Record.SAN,
		update.FEN, update.Record.WhiteClockMS, update.Record.BlackClockMS)
	if err != nil {
		return Game{}, fmt.Errorf("insert game move: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Game{}, fmt.Errorf("commit game move: %w", err)
	}
	update.Record.CreatedAt = game.UpdatedAt
	game.Moves = []MoveRecord{update.Record}
	return game, nil
}

func (r *SQLRepository) Resign(ctx context.Context, gameID, userID string) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, fmt.Errorf("begin resignation: %w", err)
	}
	defer tx.Rollback()
	var game Game
	var startedAt, finishedAt sql.NullTime
	err = tx.QueryRowContext(ctx, `
		UPDATE games SET status = 'finished',
			result = CASE WHEN white_player_id = $2 THEN '0-1' ELSE '1-0' END,
			end_reason = 'resignation', finished_at = now(), updated_at = now()
		WHERE id = $1 AND status = 'active' AND (white_player_id = $2 OR black_player_id = $2)
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, end_reason, COALESCE(draw_offer_by::text, ''),
			created_at, started_at, finished_at, updated_at`, gameID, userID,
	).Scan(&game.ID, &game.WhitePlayerID, &game.BlackPlayerID, &game.TimeControl.InitialSeconds,
		&game.TimeControl.IncrementSeconds, &game.Mode, &game.Rated, &game.CurrentFEN, &game.Status, &game.Result, &game.EndReason,
		&game.DrawOfferBy, &game.CreatedAt, &startedAt, &finishedAt, &game.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrGameNotActive
	}
	if err != nil {
		return Game{}, fmt.Errorf("resign game: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Game{}, fmt.Errorf("commit resignation: %w", err)
	}
	game.StartedAt, game.FinishedAt = nullTimePtr(startedAt), nullTimePtr(finishedAt)
	return game, nil
}

func (r *SQLRepository) OfferDraw(ctx context.Context, gameID, userID string) (Game, error) {
	game, err := scanGame(r.db.QueryRowContext(ctx, `
		UPDATE games SET draw_offer_by = $2, updated_at = now()
		WHERE id = $1 AND status = 'active' AND draw_offer_by IS NULL
		  AND (white_player_id = $2 OR black_player_id = $2)
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
			created_at, started_at, finished_at, updated_at`, gameID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrGameNotActive
	}
	if err != nil {
		return Game{}, fmt.Errorf("offer draw: %w", err)
	}
	return game, nil
}

func (r *SQLRepository) AcceptDraw(ctx context.Context, gameID, userID string) (Game, error) {
	game, err := scanGame(r.db.QueryRowContext(ctx, `
		UPDATE games SET status = 'finished', result = '1/2-1/2', end_reason = 'agreement',
			draw_offer_by = NULL, finished_at = now(), updated_at = now()
		WHERE id = $1 AND status = 'active' AND draw_offer_by IS NOT NULL AND draw_offer_by <> $2
		  AND (white_player_id = $2 OR black_player_id = $2)
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
			created_at, started_at, finished_at, updated_at`, gameID, userID))
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrDrawUnavailable
	}
	if err != nil {
		return Game{}, fmt.Errorf("accept draw: %w", err)
	}
	return game, nil
}

const gameSelect = `
	SELECT id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
	       time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
	       status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
	       created_at, started_at, finished_at, updated_at
	FROM games`

func scanGame(row interface{ Scan(...any) error }) (Game, error) {
	var game Game
	var startedAt, finishedAt sql.NullTime
	err := row.Scan(&game.ID, &game.WhitePlayerID, &game.BlackPlayerID,
		&game.TimeControl.InitialSeconds, &game.TimeControl.IncrementSeconds, &game.Mode, &game.Rated, &game.CurrentFEN,
		&game.Status, &game.Result, &game.EndReason, &game.DrawOfferBy,
		&game.CreatedAt, &startedAt, &finishedAt, &game.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrNotFound
	}
	if err != nil {
		return Game{}, fmt.Errorf("read game: %w", err)
	}
	game.StartedAt, game.FinishedAt = nullTimePtr(startedAt), nullTimePtr(finishedAt)
	game.Moves = []MoveRecord{}
	return game, nil
}

func nullTimePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}

func newGameID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("generate game ID: %w", err)
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16]), nil
}
