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
	ClaimDraw(context.Context, string, string, string) (Game, error)
	ExpireDueGames(context.Context, int) ([]Game, error)
	ClaimCompletionJobs(context.Context, int) ([]string, error)
	CompleteCompletionJob(context.Context, string) error
	RetryCompletionJob(context.Context, string, error) error
}

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func enqueueCompletionTx(ctx context.Context, tx *sql.Tx, gameID string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO game_completion_outbox (game_id) VALUES ($1)
		ON CONFLICT (game_id) DO NOTHING`, gameID)
	if err != nil {
		return fmt.Errorf("enqueue game completion: %w", err)
	}
	return nil
}

func (r *SQLRepository) ClaimCompletionJobs(ctx context.Context, limit int) ([]string, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		WITH due AS (
			SELECT game_id FROM game_completion_outbox
			WHERE available_at <= clock_timestamp()
			  AND (locked_until IS NULL OR locked_until <= clock_timestamp())
			ORDER BY available_at, game_id LIMIT $1 FOR UPDATE SKIP LOCKED
		)
		UPDATE game_completion_outbox AS jobs
		SET locked_until = clock_timestamp() + INTERVAL '30 seconds', attempts = attempts + 1
		FROM due WHERE jobs.game_id = due.game_id
		RETURNING jobs.game_id::text`, limit)
	if err != nil {
		return nil, fmt.Errorf("claim game completion jobs: %w", err)
	}
	defer rows.Close()
	var gameIDs []string
	for rows.Next() {
		var gameID string
		if err := rows.Scan(&gameID); err != nil {
			return nil, fmt.Errorf("scan game completion job: %w", err)
		}
		gameIDs = append(gameIDs, gameID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate game completion jobs: %w", err)
	}
	return gameIDs, nil
}

func (r *SQLRepository) CompleteCompletionJob(ctx context.Context, gameID string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM game_completion_outbox WHERE game_id = $1`, gameID); err != nil {
		return fmt.Errorf("complete game completion job: %w", err)
	}
	return nil
}

func (r *SQLRepository) RetryCompletionJob(ctx context.Context, gameID string, jobErr error) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE game_completion_outbox
		SET locked_until = NULL, last_error = $2,
			available_at = clock_timestamp() + LEAST(300::double precision,
				power(2::double precision, LEAST(attempts, 8))) * INTERVAL '1 second'
		WHERE game_id = $1`, gameID, jobErr.Error())
	if err != nil {
		return fmt.Errorf("retry game completion job: %w", err)
	}
	return nil
}

func clockRemaining(game Game, now time.Time) (chess.Color, int64, error) {
	if game.ClockUpdatedAt == nil {
		return chess.NoColor, 0, errors.New("active game has no clock timestamp")
	}
	position, err := chess.ParseFEN(game.CurrentFEN)
	if err != nil {
		return chess.NoColor, 0, fmt.Errorf("parse game position for clock: %w", err)
	}
	color := position.SideToMove
	remaining := game.WhiteClockMS
	if color == chess.Black {
		remaining = game.BlackClockMS
	}
	elapsed := now.Sub(*game.ClockUpdatedAt).Milliseconds()
	if elapsed < 0 {
		elapsed = 0
	}
	return color, remaining - elapsed, nil
}

func finalizeExpiredClock(ctx context.Context, tx *sql.Tx, game *Game, now time.Time) (bool, error) {
	if game.Status != StatusActive {
		return false, nil
	}
	color, remaining, err := clockRemaining(*game, now)
	if err != nil {
		return false, err
	}
	if remaining > 0 {
		return false, nil
	}
	position, err := chess.ParseFEN(game.CurrentFEN)
	if err != nil {
		return false, fmt.Errorf("parse game position on timeout: %w", err)
	}
	result, reason := ResultOnTimeout(position, color)
	if color == chess.White {
		game.WhiteClockMS = 0
	} else {
		game.BlackClockMS = 0
	}
	game.Status, game.Result, game.EndReason = StatusFinished, result, reason
	game.DrawOfferBy = ""
	game.FinishedAt, game.UpdatedAt, game.ClockUpdatedAt = &now, now, &now
	updated, err := tx.ExecContext(ctx, `
		UPDATE games SET status = 'finished', result = $2, end_reason = $3,
			finished_at = $4, updated_at = $4, clock_updated_at = $4,
			white_clock_ms = $5, black_clock_ms = $6, draw_offer_by = NULL
		WHERE id = $1 AND status = 'active'`,
		game.ID, result, reason, now, game.WhiteClockMS, game.BlackClockMS)
	if err != nil {
		return false, fmt.Errorf("finish game on timeout: %w", err)
	}
	count, err := updated.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("check timeout update: %w", err)
	}
	if count != 1 {
		return false, ErrGameNotActive
	}
	if err := enqueueCompletionTx(ctx, tx, game.ID); err != nil {
		return false, err
	}
	return true, nil
}

func (r *SQLRepository) ExpireDueGames(ctx context.Context, limit int) ([]Game, error) {
	if limit <= 0 || limit > 1000 {
		limit = 100
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin game clock expiry batch: %w", err)
	}
	defer tx.Rollback()
	rows, err := tx.QueryContext(ctx, gameSelect+`
		WHERE status = 'active' AND clock_updated_at IS NOT NULL
		  AND clock_updated_at + ((CASE WHEN split_part(current_fen, ' ', 2) = 'w'
		      THEN white_clock_ms ELSE black_clock_ms END)::double precision * INTERVAL '1 millisecond') <= clock_timestamp()
		ORDER BY clock_updated_at, id LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, fmt.Errorf("select expired game clocks: %w", err)
	}
	var due []Game
	for rows.Next() {
		game, err := scanGame(rows)
		if err != nil {
			_ = rows.Close()
			return nil, fmt.Errorf("scan expired game: %w", err)
		}
		due = append(due, game)
	}
	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return nil, fmt.Errorf("iterate expired games: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close expired game rows: %w", err)
	}
	expired := make([]Game, 0, len(due))
	for _, game := range due {
		var now time.Time
		if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
			return nil, fmt.Errorf("read timeout timestamp: %w", err)
		}
		finalized, err := finalizeExpiredClock(ctx, tx, &game, now)
		if err != nil {
			return nil, err
		}
		if finalized {
			expired = append(expired, game)
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit game clock expiry batch: %w", err)
	}
	return expired, nil
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
			mode, rated, current_fen, status, result, white_clock_ms, black_clock_ms
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, '*', $9, $9)
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
			white_clock_ms, black_clock_ms, clock_updated_at,
			created_at, started_at, finished_at, updated_at`,
		id, userID, input.TimeControl.InitialSeconds, input.TimeControl.IncrementSeconds,
		input.Mode, input.Rated, chess.StartingPosition().FEN(), status, int64(input.TimeControl.InitialSeconds)*1000))
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
			white_clock_ms = time_control_initial_seconds::BIGINT * 1000,
			black_clock_ms = time_control_initial_seconds::BIGINT * 1000,
			clock_updated_at = clock_timestamp(), started_at = now(), updated_at = now()
		WHERE id = $1 AND status = 'waiting' AND black_player_id IS NULL
		RETURNING id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
			time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
			status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
			white_clock_ms, black_clock_ms, clock_updated_at,
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

	game, err := scanGame(tx.QueryRowContext(ctx, gameSelect+` WHERE id = $1 FOR UPDATE`, gameID))
	if err != nil {
		return Game{}, err
	}
	if game.CurrentFEN != expectedFEN || game.Status != StatusActive ||
		(game.WhitePlayerID != userID && game.BlackPlayerID != userID) {
		return Game{}, ErrMoveConflict
	}
	var now time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Game{}, fmt.Errorf("read game clock: %w", err)
	}
	color, remaining, err := clockRemaining(game, now)
	if err != nil {
		return Game{}, err
	}
	if (color == chess.White && game.WhitePlayerID != userID) || (color == chess.Black && game.BlackPlayerID != userID) {
		return Game{}, ErrNotYourTurn
	}
	expired, err := finalizeExpiredClock(ctx, tx, &game, now)
	if err != nil {
		return Game{}, err
	}
	if expired {
		if err := tx.Commit(); err != nil {
			return Game{}, fmt.Errorf("commit game timeout: %w", err)
		}
		return game, ErrTimeExpired
	}

	remaining += int64(game.TimeControl.IncrementSeconds) * 1000
	if userID == game.WhitePlayerID {
		game.WhiteClockMS = remaining
	} else {
		game.BlackClockMS = remaining
	}
	update.Record.WhiteClockMS = int64Ptr(game.WhiteClockMS)
	update.Record.BlackClockMS = int64Ptr(game.BlackClockMS)
	game.CurrentFEN, game.Status, game.Result, game.EndReason = update.FEN, update.Status, update.Result, update.Reason
	game.DrawOfferBy = ""
	game.UpdatedAt, game.ClockUpdatedAt = now, &now
	if update.Status == StatusFinished {
		game.FinishedAt = &now
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE games SET current_fen = $2, status = $3, result = $4, end_reason = NULLIF($5, ''),
			finished_at = CASE WHEN $3 = 'finished' THEN $6 ELSE NULL END,
			white_clock_ms = $7, black_clock_ms = $8, clock_updated_at = $6, updated_at = $6,
			draw_offer_by = NULL
		WHERE id = $1`, gameID, update.FEN, update.Status, update.Result, update.Reason, now,
		game.WhiteClockMS, game.BlackClockMS); err != nil {
		return Game{}, fmt.Errorf("update game position: %w", err)
	}
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
	if update.Status == StatusFinished {
		if err := enqueueCompletionTx(ctx, tx, gameID); err != nil {
			return Game{}, err
		}
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
	game, err := scanGame(tx.QueryRowContext(ctx, gameSelect+` WHERE id = $1 FOR UPDATE`, gameID))
	if err != nil {
		return Game{}, err
	}
	if game.Status != StatusActive {
		return Game{}, ErrGameNotActive
	}
	if game.WhitePlayerID != userID && game.BlackPlayerID != userID {
		return Game{}, ErrForbidden
	}
	var now time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Game{}, fmt.Errorf("read resignation timestamp: %w", err)
	}
	expired, err := finalizeExpiredClock(ctx, tx, &game, now)
	if err != nil {
		return Game{}, err
	}
	if expired {
		if err := tx.Commit(); err != nil {
			return Game{}, fmt.Errorf("commit resignation timeout: %w", err)
		}
		return game, ErrTimeExpired
	}
	result := ResultWhiteWin
	if userID == game.WhitePlayerID {
		result = ResultBlackWin
	}
	color, remaining, err := clockRemaining(game, now)
	if err != nil {
		return Game{}, err
	}
	if color == chess.White {
		game.WhiteClockMS = remaining
	} else {
		game.BlackClockMS = remaining
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE games SET status = 'finished', result = $2, end_reason = 'resignation',
			finished_at = $3, updated_at = $3, clock_updated_at = $3, draw_offer_by = NULL,
			white_clock_ms = $4, black_clock_ms = $5
		WHERE id = $1`, gameID, result, now, game.WhiteClockMS, game.BlackClockMS); err != nil {
		return Game{}, fmt.Errorf("resign game: %w", err)
	}
	game.Status, game.Result, game.EndReason = StatusFinished, result, "resignation"
	game.DrawOfferBy = ""
	game.FinishedAt, game.UpdatedAt, game.ClockUpdatedAt = &now, now, &now
	if err := enqueueCompletionTx(ctx, tx, gameID); err != nil {
		return Game{}, err
	}
	if err := tx.Commit(); err != nil {
		return Game{}, fmt.Errorf("commit resignation: %w", err)
	}
	return game, nil
}

func (r *SQLRepository) OfferDraw(ctx context.Context, gameID, userID string) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, fmt.Errorf("begin draw offer: %w", err)
	}
	defer tx.Rollback()
	game, err := scanGame(tx.QueryRowContext(ctx, gameSelect+` WHERE id = $1 FOR UPDATE`, gameID))
	if err != nil {
		return Game{}, err
	}
	if game.Status != StatusActive {
		return Game{}, ErrGameNotActive
	}
	if game.WhitePlayerID != userID && game.BlackPlayerID != userID {
		return Game{}, ErrForbidden
	}
	var now time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Game{}, fmt.Errorf("read draw offer timestamp: %w", err)
	}
	expired, err := finalizeExpiredClock(ctx, tx, &game, now)
	if err != nil {
		return Game{}, err
	}
	if expired {
		if err := tx.Commit(); err != nil {
			return Game{}, fmt.Errorf("commit draw offer timeout: %w", err)
		}
		return game, ErrTimeExpired
	}
	if game.DrawOfferBy != "" {
		return Game{}, ErrDrawUnavailable
	}
	if _, err := tx.ExecContext(ctx, `UPDATE games SET draw_offer_by = $2, updated_at = $3 WHERE id = $1`, gameID, userID, now); err != nil {
		return Game{}, fmt.Errorf("offer draw: %w", err)
	}
	game.DrawOfferBy, game.UpdatedAt = userID, now
	if err := tx.Commit(); err != nil {
		return Game{}, fmt.Errorf("commit draw offer: %w", err)
	}
	return game, nil
}

func (r *SQLRepository) AcceptDraw(ctx context.Context, gameID, userID string) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, fmt.Errorf("begin draw acceptance: %w", err)
	}
	defer tx.Rollback()
	game, err := scanGame(tx.QueryRowContext(ctx, gameSelect+` WHERE id = $1 FOR UPDATE`, gameID))
	if err != nil {
		return Game{}, err
	}
	if game.Status != StatusActive {
		return Game{}, ErrGameNotActive
	}
	if game.WhitePlayerID != userID && game.BlackPlayerID != userID {
		return Game{}, ErrForbidden
	}
	var now time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Game{}, fmt.Errorf("read draw acceptance timestamp: %w", err)
	}
	expired, err := finalizeExpiredClock(ctx, tx, &game, now)
	if err != nil {
		return Game{}, err
	}
	if expired {
		if err := tx.Commit(); err != nil {
			return Game{}, fmt.Errorf("commit draw acceptance timeout: %w", err)
		}
		return game, ErrTimeExpired
	}
	if game.DrawOfferBy == "" || game.DrawOfferBy == userID {
		return Game{}, ErrDrawUnavailable
	}
	color, remaining, err := clockRemaining(game, now)
	if err != nil {
		return Game{}, err
	}
	if color == chess.White {
		game.WhiteClockMS = remaining
	} else {
		game.BlackClockMS = remaining
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE games SET status = 'finished', result = '1/2-1/2', end_reason = 'agreement',
			draw_offer_by = NULL, finished_at = $2, updated_at = $2, clock_updated_at = $2,
			white_clock_ms = $3, black_clock_ms = $4
		WHERE id = $1`, gameID, now, game.WhiteClockMS, game.BlackClockMS); err != nil {
		return Game{}, fmt.Errorf("accept draw: %w", err)
	}
	game.Status, game.Result, game.EndReason, game.DrawOfferBy = StatusFinished, ResultDraw, "agreement", ""
	game.FinishedAt, game.UpdatedAt, game.ClockUpdatedAt = &now, now, &now
	if err := enqueueCompletionTx(ctx, tx, gameID); err != nil {
		return Game{}, err
	}
	if err := tx.Commit(); err != nil {
		return Game{}, fmt.Errorf("commit draw acceptance: %w", err)
	}
	return game, nil
}

func (r *SQLRepository) ClaimDraw(ctx context.Context, gameID, userID, reason string) (Game, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Game{}, fmt.Errorf("begin draw claim: %w", err)
	}
	defer tx.Rollback()
	game, err := scanGame(tx.QueryRowContext(ctx, gameSelect+` WHERE id = $1 FOR UPDATE`, gameID))
	if err != nil {
		return Game{}, err
	}
	if game.Status != StatusActive {
		return Game{}, ErrGameNotActive
	}
	if game.WhitePlayerID != userID && game.BlackPlayerID != userID {
		return Game{}, ErrForbidden
	}
	position, err := chess.ParseFEN(game.CurrentFEN)
	if err != nil {
		return Game{}, fmt.Errorf("parse position for draw claim: %w", err)
	}
	var now time.Time
	if err := tx.QueryRowContext(ctx, `SELECT clock_timestamp()`).Scan(&now); err != nil {
		return Game{}, fmt.Errorf("read draw claim timestamp: %w", err)
	}
	color, remaining, err := clockRemaining(game, now)
	if err != nil {
		return Game{}, err
	}
	if (color == chess.White && game.WhitePlayerID != userID) || (color == chess.Black && game.BlackPlayerID != userID) {
		return Game{}, ErrNotYourTurn
	}
	expired, err := finalizeExpiredClock(ctx, tx, &game, now)
	if err != nil {
		return Game{}, err
	}
	if expired {
		if err := tx.Commit(); err != nil {
			return Game{}, fmt.Errorf("commit draw claim timeout: %w", err)
		}
		return game, ErrTimeExpired
	}
	claimValid := false
	endReason := ""
	switch reason {
	case "fifty_move_rule":
		claimValid = position.IsDrawByFiftyMoveRule()
		endReason = "fifty_move_rule_claim"
	case "threefold_repetition":
		var count int
		count, err = countPositionRepetitions(ctx, tx, game.ID, position)
		if err != nil {
			return Game{}, err
		}
		claimValid = count >= 3
		endReason = "threefold_repetition_claim"
	default:
		return Game{}, ErrInvalidRequest
	}
	if !claimValid {
		return Game{}, ErrDrawUnavailable
	}
	if color == chess.White {
		game.WhiteClockMS = remaining
	} else {
		game.BlackClockMS = remaining
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE games SET status = 'finished', result = '1/2-1/2', end_reason = $2,
			finished_at = $3, updated_at = $3, clock_updated_at = $3,
			white_clock_ms = $4, black_clock_ms = $5, draw_offer_by = NULL
		WHERE id = $1`, gameID, endReason, now, game.WhiteClockMS, game.BlackClockMS); err != nil {
		return Game{}, fmt.Errorf("apply draw claim: %w", err)
	}
	game.Status, game.Result, game.EndReason, game.DrawOfferBy = StatusFinished, ResultDraw, endReason, ""
	game.FinishedAt, game.UpdatedAt, game.ClockUpdatedAt = &now, now, &now
	if err := enqueueCompletionTx(ctx, tx, gameID); err != nil {
		return Game{}, err
	}
	if err := tx.Commit(); err != nil {
		return Game{}, fmt.Errorf("commit draw claim: %w", err)
	}
	return game, nil
}

func countPositionRepetitions(ctx context.Context, tx *sql.Tx, gameID string, target chess.Position) (int, error) {
	key := target.RepetitionKey()
	count := 0
	if chess.StartingPosition().RepetitionKey() == key {
		count++
	}
	rows, err := tx.QueryContext(ctx, `SELECT fen_after FROM game_moves WHERE game_id = $1 ORDER BY ply`, gameID)
	if err != nil {
		return 0, fmt.Errorf("load game history for draw claim: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var fen string
		if err := rows.Scan(&fen); err != nil {
			return 0, fmt.Errorf("scan game history for draw claim: %w", err)
		}
		position, err := chess.ParseFEN(fen)
		if err != nil {
			return 0, fmt.Errorf("parse game history for draw claim: %w", err)
		}
		if position.RepetitionKey() == key {
			count++
		}
	}
	if err := rows.Err(); err != nil {
		return 0, fmt.Errorf("iterate game history for draw claim: %w", err)
	}
	return count, nil
}

const gameSelect = `
	SELECT id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
	       time_control_initial_seconds, time_control_increment_seconds, mode, rated, current_fen,
	       status, result, COALESCE(end_reason, ''), COALESCE(draw_offer_by::text, ''),
	       white_clock_ms, black_clock_ms, clock_updated_at,
	       created_at, started_at, finished_at, updated_at
	FROM games`

func scanGame(row interface{ Scan(...any) error }) (Game, error) {
	var game Game
	var clockUpdatedAt, startedAt, finishedAt sql.NullTime
	err := row.Scan(&game.ID, &game.WhitePlayerID, &game.BlackPlayerID,
		&game.TimeControl.InitialSeconds, &game.TimeControl.IncrementSeconds, &game.Mode, &game.Rated, &game.CurrentFEN,
		&game.Status, &game.Result, &game.EndReason, &game.DrawOfferBy,
		&game.WhiteClockMS, &game.BlackClockMS, &clockUpdatedAt,
		&game.CreatedAt, &startedAt, &finishedAt, &game.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Game{}, ErrNotFound
	}
	if err != nil {
		return Game{}, fmt.Errorf("read game: %w", err)
	}
	game.StartedAt, game.FinishedAt = nullTimePtr(startedAt), nullTimePtr(finishedAt)
	game.ClockUpdatedAt = nullTimePtr(clockUpdatedAt)
	game.Moves = []MoveRecord{}
	return game, nil
}



func int64Ptr(value int64) *int64 { return &value }

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
