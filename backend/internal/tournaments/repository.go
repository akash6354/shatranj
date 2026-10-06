package tournaments

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List(ctx context.Context) ([]Tournament, error) {
	rows, err := r.db.QueryContext(ctx, tournamentColumns+`
		WHERE status IN ('upcoming', 'live')
		ORDER BY starts_at, id`)
	if err != nil {
		return nil, fmt.Errorf("list tournaments: %w", err)
	}
	defer rows.Close()
	result := make([]Tournament, 0)
	for rows.Next() {
		tournament, err := scanTournament(rows)
		if err != nil {
			return nil, fmt.Errorf("scan tournament: %w", err)
		}
		result = append(result, tournament)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tournaments: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) Create(ctx context.Context, creator string, input CreateInput) (Tournament, error) {
	return scanTournament(r.db.QueryRowContext(ctx, `
		INSERT INTO tournaments (creator_id, name, description, format, time_control_seconds,
			increment_seconds, max_players, starts_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id::text, creator_id::text, name, description, format, status,
			time_control_seconds, increment_seconds, max_players, starts_at, created_at, updated_at`,
		creator, input.Name, input.Description, input.Format, input.TimeControlSecs,
		input.IncrementSeconds, input.MaxPlayers, input.StartsAt))
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Tournament, error) {
	tournament, err := scanTournament(r.db.QueryRowContext(ctx, tournamentColumns+` WHERE id = $1`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Tournament{}, ErrNotFound
	}
	if err != nil {
		return Tournament{}, fmt.Errorf("get tournament: %w", err)
	}
	tournament.Participants, err = r.participants(ctx, id)
	if err != nil {
		return Tournament{}, err
	}
	tournament.Rounds, err = r.rounds(ctx, id)
	if err != nil {
		return Tournament{}, err
	}
	return tournament, nil
}

func (r *PostgresRepository) Register(ctx context.Context, id, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tournament registration: %w", err)
	}
	defer tx.Rollback()
	var status Status
	var maxPlayers int
	if err := tx.QueryRowContext(ctx, `SELECT status, max_players FROM tournaments WHERE id = $1 FOR UPDATE`, id).
		Scan(&status, &maxPlayers); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lock tournament for registration: %w", err)
	}
	if status != StatusUpcoming {
		return ErrConflict
	}
	var already bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM tournament_participants WHERE tournament_id = $1 AND user_id = $2)`,
		id, userID).Scan(&already); err != nil {
		return fmt.Errorf("check tournament participant: %w", err)
	}
	if already {
		return ErrConflict
	}
	var count int
	if err := tx.QueryRowContext(ctx, `SELECT count(*) FROM tournament_participants WHERE tournament_id = $1`, id).Scan(&count); err != nil {
		return fmt.Errorf("count tournament participants: %w", err)
	}
	if count >= maxPlayers {
		return ErrConflict
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tournament_participants (tournament_id, user_id) VALUES ($1, $2)`, id, userID); err != nil {
		return fmt.Errorf("register for tournament: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tournament registration: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Withdraw(ctx context.Context, id, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		DELETE FROM tournament_participants p USING tournaments t
		WHERE p.tournament_id = t.id AND t.id = $1 AND p.user_id = $2 AND t.status = 'upcoming'`, id, userID)
	if err != nil {
		return fmt.Errorf("withdraw from tournament: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read withdrawal result: %w", err)
	}
	if affected == 0 {
		return r.registrationError(ctx, id)
	}
	return nil
}

func (r *PostgresRepository) StartRound(ctx context.Context, id, userID string) (Round, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Round{}, fmt.Errorf("begin tournament round: %w", err)
	}
	defer tx.Rollback()
	var creator string
	var status Status
	if err := tx.QueryRowContext(ctx, `
		SELECT creator_id::text, status FROM tournaments WHERE id = $1 FOR UPDATE`, id).Scan(&creator, &status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Round{}, ErrNotFound
		}
		return Round{}, fmt.Errorf("lock tournament: %w", err)
	}
	if creator != userID {
		return Round{}, ErrForbidden
	}
	if status != StatusUpcoming && status != StatusLive {
		return Round{}, ErrConflict
	}
	var priorIncomplete bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM tournament_rounds WHERE tournament_id = $1 AND status <> 'completed')`, id).Scan(&priorIncomplete); err != nil {
		return Round{}, fmt.Errorf("check active tournament rounds: %w", err)
	}
	if priorIncomplete {
		return Round{}, ErrConflict
	}
	rows, err := tx.QueryContext(ctx, `
		SELECT user_id::text, rating FROM tournament_participants
		WHERE tournament_id = $1 ORDER BY score DESC, rating DESC, user_id`, id)
	if err != nil {
		return Round{}, fmt.Errorf("load tournament participants: %w", err)
	}
	players := make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.UserID, &player.Rating); err != nil {
			rows.Close()
			return Round{}, fmt.Errorf("scan tournament participant: %w", err)
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Round{}, fmt.Errorf("iterate tournament participants: %w", err)
	}
	if err := rows.Close(); err != nil {
		return Round{}, fmt.Errorf("close tournament participants: %w", err)
	}
	if len(players) < 2 {
		return Round{}, ErrConflict
	}
	var number int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(max(round_number), 0) + 1 FROM tournament_rounds WHERE tournament_id = $1`, id).Scan(&number); err != nil {
		return Round{}, fmt.Errorf("get next tournament round: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO tournament_rounds (tournament_id, round_number) VALUES ($1, $2)`, id, number); err != nil {
		return Round{}, fmt.Errorf("create tournament round: %w", err)
	}
	pairings := PairPlayers(players)
	for index := range pairings {
		pairing := &pairings[index]
		err := tx.QueryRowContext(ctx, `
			INSERT INTO tournament_pairings
				(tournament_id, round_number, white_player_id, black_player_id, result, white_score, black_score)
			VALUES ($1, $2, $3, NULLIF($4, '')::uuid, $5, $6, $7)
			RETURNING id::text`,
			id, number, pairing.WhiteID, pairing.BlackID, pairing.Result, pairing.WhiteScore, pairing.BlackScore).
			Scan(&pairing.ID)
		if err != nil {
			return Round{}, fmt.Errorf("create tournament pairing: %w", err)
		}
		if pairing.Bye {
			if _, err := tx.ExecContext(ctx, `
				UPDATE tournament_participants SET score = score + $3
				WHERE tournament_id = $1 AND user_id = $2`, id, pairing.WhiteID, pairing.WhiteScore); err != nil {
				return Round{}, fmt.Errorf("award tournament bye: %w", err)
			}
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE tournaments SET status = 'live', updated_at = now() WHERE id = $1`, id); err != nil {
		return Round{}, fmt.Errorf("start tournament: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Round{}, fmt.Errorf("commit tournament round: %w", err)
	}
	return Round{Number: number, Status: "in_progress", Pairings: pairings}, nil
}

func (r *PostgresRepository) Standings(ctx context.Context, id string) ([]Player, error) {
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM tournaments WHERE id = $1)`, id).Scan(&exists); err != nil {
		return nil, fmt.Errorf("check tournament for standings: %w", err)
	}
	if !exists {
		return nil, ErrNotFound
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.user_id::text, COALESCE(u.username, ''), p.rating, p.score
		FROM tournament_participants p
		JOIN tournaments t ON t.id = p.tournament_id
		LEFT JOIN profiles u ON u.user_id = p.user_id
		WHERE p.tournament_id = $1
		ORDER BY p.score DESC, p.rating DESC, p.user_id`, id)
	if err != nil {
		return nil, fmt.Errorf("get tournament standings: %w", err)
	}

	defer rows.Close()
	players := make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.UserID, &player.Username, &player.Rating, &player.Score); err != nil {
			return nil, fmt.Errorf("scan tournament standing: %w", err)
		}
		players = append(players, player)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tournament standings: %w", err)
	}
	SortStandings(players)
	return players, nil
}

func (r *PostgresRepository) ReportResult(ctx context.Context, tournamentID, pairingID, userID, result string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tournament result: %w", err)
	}
	defer tx.Rollback()

	var creator string
	if err := tx.QueryRowContext(ctx, `
		SELECT creator_id::text FROM tournaments WHERE id = $1 FOR UPDATE`, tournamentID).Scan(&creator); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNotFound
		}
		return fmt.Errorf("lock tournament for result: %w", err)
	}
	if creator != userID {
		return ErrForbidden
	}

	var round int
	var whiteID, blackID string
	err = tx.QueryRowContext(ctx, `
		SELECT round_number, white_player_id::text, COALESCE(black_player_id::text, '')
		FROM tournament_pairings
		WHERE id = $1 AND tournament_id = $2 AND result = '*' FOR UPDATE`,
		pairingID, tournamentID).Scan(&round, &whiteID, &blackID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrConflict
	}
	if err != nil {
		return fmt.Errorf("lock tournament pairing: %w", err)
	}
	if blackID == "" {
		return ErrConflict
	}

	whiteScore, blackScore := 0, 0
	switch result {
	case "1-0":
		whiteScore = 2
	case "0-1":
		blackScore = 2
	case "1/2-1/2":
		whiteScore, blackScore = 1, 1
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tournament_pairings SET result = $3, white_score = $4, black_score = $5
		WHERE id = $1 AND tournament_id = $2`, pairingID, tournamentID, result, whiteScore, blackScore); err != nil {
		return fmt.Errorf("update tournament pairing result: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tournament_participants SET score = score + CASE user_id WHEN $2 THEN $4 ELSE $5 END
		WHERE tournament_id = $1 AND user_id IN ($2, $3)`, tournamentID, whiteID, blackID, whiteScore, blackScore); err != nil {
		return fmt.Errorf("update tournament standings scores: %w", err)
	}

	var pending bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM tournament_pairings WHERE tournament_id = $1 AND round_number = $2 AND result = '*')`,
		tournamentID, round).Scan(&pending); err != nil {
		return fmt.Errorf("check tournament round completion: %w", err)
	}
	if !pending {
		if _, err := tx.ExecContext(ctx, `
			UPDATE tournament_rounds SET status = 'completed', completed_at = now()
			WHERE tournament_id = $1 AND round_number = $2`, tournamentID, round); err != nil {
			return fmt.Errorf("complete tournament round: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit tournament result: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Complete(ctx context.Context, tournamentID, userID string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE tournaments t SET status = 'completed', updated_at = now()
		WHERE t.id = $1 AND t.creator_id = $2 AND t.status = 'live'
		  AND NOT EXISTS (SELECT 1 FROM tournament_rounds r WHERE r.tournament_id = t.id AND r.status <> 'completed')`,
		tournamentID, userID)
	if err != nil {
		return fmt.Errorf("complete tournament: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read tournament completion result: %w", err)
	}
	if affected == 1 {
		return nil
	}

	var exists, owner bool
	if err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM tournaments WHERE id = $1),
			EXISTS(SELECT 1 FROM tournaments WHERE id = $1 AND creator_id = $2)`,
		tournamentID, userID).Scan(&exists, &owner); err != nil {
		return fmt.Errorf("check tournament completion: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	if !owner {
		return ErrForbidden
	}
	return ErrConflict
}

func (r *PostgresRepository) participants(ctx context.Context, id string) ([]Player, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.user_id::text, COALESCE(u.username, ''), p.rating, p.score
		FROM tournament_participants p LEFT JOIN profiles u ON u.user_id = p.user_id
		WHERE p.tournament_id = $1 ORDER BY p.created_at, p.user_id`, id)
	if err != nil {
		return nil, fmt.Errorf("list tournament participants: %w", err)
	}
	defer rows.Close()
	result := make([]Player, 0)
	for rows.Next() {
		var player Player
		if err := rows.Scan(&player.UserID, &player.Username, &player.Rating, &player.Score); err != nil {
			return nil, fmt.Errorf("scan tournament player: %w", err)
		}
		result = append(result, player)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tournament players: %w", err)
	}
	return result, nil
}

func (r *PostgresRepository) rounds(ctx context.Context, id string) ([]Round, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT round_number, status FROM tournament_rounds
		WHERE tournament_id = $1 ORDER BY round_number`, id)
	if err != nil {
		return nil, fmt.Errorf("list tournament rounds: %w", err)
	}
	result := make([]Round, 0)
	for rows.Next() {
		var round Round
		if err := rows.Scan(&round.Number, &round.Status); err != nil {
			rows.Close()
			return nil, fmt.Errorf("scan tournament round: %w", err)
		}
		result = append(result, round)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, fmt.Errorf("iterate tournament rounds: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("close tournament rounds: %w", err)
	}

	for index := range result {
		pairRows, err := r.db.QueryContext(ctx, `
			SELECT id::text, white_player_id::text, COALESCE(black_player_id::text, ''),
				COALESCE(game_id::text, ''), result, white_score, black_score,
				black_player_id IS NULL
			FROM tournament_pairings WHERE tournament_id = $1 AND round_number = $2
			ORDER BY id`, id, result[index].Number)
		if err != nil {
			return nil, fmt.Errorf("list tournament pairings: %w", err)
		}
		result[index].Pairings = make([]Pairing, 0)
		for pairRows.Next() {
			var pairing Pairing
			if err := pairRows.Scan(&pairing.ID, &pairing.WhiteID, &pairing.BlackID,
				&pairing.GameID, &pairing.Result, &pairing.WhiteScore, &pairing.BlackScore, &pairing.Bye); err != nil {
				pairRows.Close()
				return nil, fmt.Errorf("scan tournament pairing: %w", err)
			}
			result[index].Pairings = append(result[index].Pairings, pairing)
		}
		if err := pairRows.Err(); err != nil {
			pairRows.Close()
			return nil, fmt.Errorf("iterate tournament pairings: %w", err)
		}
		if err := pairRows.Close(); err != nil {
			return nil, fmt.Errorf("close tournament pairings: %w", err)
		}
	}
	return result, nil
}

func (r *PostgresRepository) registrationError(ctx context.Context, id string) error {
	var exists bool
	if err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM tournaments WHERE id = $1)`, id).Scan(&exists); err != nil {
		return fmt.Errorf("check tournament registration: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	return ErrConflict
}

const tournamentColumns = `SELECT id::text, creator_id::text, name, description, format, status,
	time_control_seconds, increment_seconds, max_players, starts_at, created_at, updated_at FROM tournaments`

func scanTournament(row interface{ Scan(...any) error }) (Tournament, error) {
	var tournament Tournament
	err := row.Scan(&tournament.ID, &tournament.CreatorID, &tournament.Name, &tournament.Description,
		&tournament.Format, &tournament.Status, &tournament.TimeControlSecs, &tournament.IncrementSeconds,
		&tournament.MaxPlayers, &tournament.StartsAt, &tournament.CreatedAt, &tournament.UpdatedAt)
	return tournament, err
}

var _ Repository = (*PostgresRepository)(nil)
