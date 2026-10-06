package clubs

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) List(ctx context.Context, userID string) ([]Club, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT c.id::text, c.owner_id::text, c.name, c.description, c.visibility,
			COALESCE(m.role, ''), c.created_at
		FROM clubs c LEFT JOIN club_members m ON m.club_id = c.id AND m.user_id = NULLIF($1, '')::uuid
		WHERE c.visibility = 'public' OR m.user_id IS NOT NULL
		ORDER BY c.name, c.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list clubs: %w", err)
	}
	defer rows.Close()
	clubs := make([]Club, 0)
	for rows.Next() {
		var club Club
		if err := rows.Scan(&club.ID, &club.OwnerID, &club.Name, &club.Description,
			&club.Visibility, &club.Role, &club.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan club: %w", err)
		}
		clubs = append(clubs, club)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate clubs: %w", err)
	}
	return clubs, nil
}

func (r *PostgresRepository) Create(ctx context.Context, userID string, input CreateInput) (Club, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Club{}, fmt.Errorf("begin club creation: %w", err)
	}
	defer tx.Rollback()
	var club Club
	err = tx.QueryRowContext(ctx, `
		INSERT INTO clubs (owner_id, name, description, visibility)
		VALUES ($1, $2, $3, $4)
		RETURNING id::text, owner_id::text, name, description, visibility, created_at`,
		userID, input.Name, input.Description, input.Visibility).Scan(
		&club.ID, &club.OwnerID, &club.Name, &club.Description, &club.Visibility, &club.CreatedAt)
	if err != nil {
		return Club{}, fmt.Errorf("create club: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO club_members (club_id, user_id, role) VALUES ($1, $2, 'owner')`, club.ID, userID); err != nil {
		return Club{}, fmt.Errorf("add club owner membership: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Club{}, fmt.Errorf("commit club creation: %w", err)
	}
	club.Role = "owner"
	return club, nil
}

func (r *PostgresRepository) Get(ctx context.Context, clubID, userID string) (Club, error) {
	var club Club
	err := r.db.QueryRowContext(ctx, `
		SELECT c.id::text, c.owner_id::text, c.name, c.description, c.visibility,
			COALESCE(m.role, ''), c.created_at
		FROM clubs c LEFT JOIN club_members m ON m.club_id = c.id AND m.user_id = NULLIF($2, '')::uuid
		WHERE c.id = $1 AND (c.visibility = 'public' OR m.user_id IS NOT NULL)`, clubID, userID).Scan(
		&club.ID, &club.OwnerID, &club.Name, &club.Description, &club.Visibility, &club.Role, &club.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Club{}, ErrNotFound
	}
	if err != nil {
		return Club{}, fmt.Errorf("get club: %w", err)
	}
	return club, nil
}

func (r *PostgresRepository) Members(ctx context.Context, clubID, _ string) ([]Member, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.user_id::text, COALESCE(p.username, ''), m.role, m.joined_at
		FROM club_members m LEFT JOIN profiles p ON p.user_id = m.user_id
		WHERE m.club_id = $1 ORDER BY CASE m.role WHEN 'owner' THEN 0 WHEN 'admin' THEN 1 ELSE 2 END, m.joined_at`,
		clubID)
	if err != nil {
		return nil, fmt.Errorf("list club members: %w", err)
	}
	defer rows.Close()
	members := make([]Member, 0)
	for rows.Next() {
		var member Member
		if err := rows.Scan(&member.UserID, &member.Username, &member.Role, &member.JoinedAt); err != nil {
			return nil, fmt.Errorf("scan club member: %w", err)
		}
		members = append(members, member)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate club members: %w", err)
	}
	return members, nil
}

func (r *PostgresRepository) Join(ctx context.Context, clubID, userID string) (string, error) {
	var visibility string
	err := r.db.QueryRowContext(ctx, `SELECT visibility FROM clubs WHERE id = $1`, clubID).Scan(&visibility)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get club for join: %w", err)
	}
	if visibility == "public" {
		_, err = r.db.ExecContext(ctx, `
			INSERT INTO club_members (club_id, user_id, role) VALUES ($1, $2, 'member')
			ON CONFLICT (club_id, user_id) DO NOTHING`, clubID, userID)
		if err != nil {
			return "", fmt.Errorf("join public club: %w", err)
		}
		return "joined", nil
	}
	var status string
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO club_join_requests (club_id, user_id)
		VALUES ($1, $2)
		ON CONFLICT (club_id, user_id) DO UPDATE
			SET status = CASE WHEN club_join_requests.status = 'rejected' THEN 'pending' ELSE club_join_requests.status END,
				updated_at = CASE WHEN club_join_requests.status = 'rejected' THEN now() ELSE club_join_requests.updated_at END
		RETURNING status`, clubID, userID).Scan(&status)
	if err != nil {
		return "", fmt.Errorf("request private club access: %w", err)
	}
	if status == "accepted" {
		if _, err := r.db.ExecContext(ctx, `
			INSERT INTO club_members (club_id, user_id, role) VALUES ($1, $2, 'member')
			ON CONFLICT (club_id, user_id) DO NOTHING`, clubID, userID); err != nil {
			return "", fmt.Errorf("restore private club membership: %w", err)
		}
		return "joined", nil
	}
	return "pending", nil
}

func (r *PostgresRepository) Leave(ctx context.Context, clubID, userID string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin club leave: %w", err)
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		DELETE FROM club_members WHERE club_id = $1 AND user_id = $2 AND role <> 'owner'`,
		clubID, userID)
	if err != nil {
		return fmt.Errorf("leave club: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read club leave result: %w", err)
	}
	var exists bool
	if err := tx.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM clubs WHERE id = $1)`, clubID).Scan(&exists); err != nil {
		return fmt.Errorf("check club after leave: %w", err)
	}
	if !exists {
		return ErrNotFound
	}
	requestResult, err := tx.ExecContext(ctx, `
		DELETE FROM club_join_requests WHERE club_id = $1 AND user_id = $2`,
		clubID, userID)
	if err != nil {
		return fmt.Errorf("withdraw club join request: %w", err)
	}
	requests, err := requestResult.RowsAffected()
	if err != nil {
		return fmt.Errorf("read club request withdrawal result: %w", err)
	}
	if affected == 0 && requests == 0 {
		return ErrConflict
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit club leave: %w", err)
	}
	return nil
}

func (r *PostgresRepository) Requests(ctx context.Context, clubID, userID string) ([]JoinRequest, error) {
	if err := r.requireModerator(ctx, clubID, userID); err != nil {
		return nil, err
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT q.id::text, q.user_id::text, COALESCE(p.username, ''), q.status, q.created_at
		FROM club_join_requests q LEFT JOIN profiles p ON p.user_id = q.user_id
		WHERE q.club_id = $1 AND q.status = 'pending' ORDER BY q.created_at`, clubID)
	if err != nil {
		return nil, fmt.Errorf("list club join requests: %w", err)
	}
	defer rows.Close()
	requests := make([]JoinRequest, 0)
	for rows.Next() {
		var request JoinRequest
		if err := rows.Scan(&request.ID, &request.UserID, &request.Username, &request.Status, &request.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan club join request: %w", err)
		}
		requests = append(requests, request)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate club join requests: %w", err)
	}
	return requests, nil
}

func (r *PostgresRepository) ResolveRequest(ctx context.Context, clubID, requestID, adminID string, approve bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin club request resolution: %w", err)
	}
	defer tx.Rollback()
	var role string
	err = tx.QueryRowContext(ctx, `
		SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2 FOR UPDATE`, clubID, adminID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("authorize club request resolution: %w", err)
	}
	if role != "owner" && role != "admin" {
		return ErrForbidden
	}
	var target string
	err = tx.QueryRowContext(ctx, `
		UPDATE club_join_requests SET status = $3, updated_at = now()
		WHERE id = $1 AND club_id = $2 AND status = 'pending'
		RETURNING user_id::text`, requestID, clubID, acceptedOrRejected(approve)).Scan(&target)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("resolve club join request: %w", err)
	}
	if approve {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO club_members (club_id, user_id, role) VALUES ($1, $2, 'member')
			ON CONFLICT (club_id, user_id) DO NOTHING`, clubID, target); err != nil {
			return fmt.Errorf("add approved club member: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit club request resolution: %w", err)
	}
	return nil
}

func (r *PostgresRepository) SetRole(ctx context.Context, clubID, targetID, adminID, role string) error {
	if err := r.requireOwner(ctx, clubID, adminID); err != nil {
		return err
	}
	result, err := r.db.ExecContext(ctx, `
		UPDATE club_members SET role = $3 WHERE club_id = $1 AND user_id = $2 AND role <> 'owner'`,
		clubID, targetID, role)
	if err != nil {
		return fmt.Errorf("set club member role: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read role update result: %w", err)
	}
	if affected == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) Authorize(ctx context.Context, clubID, userID string) error {
	if userID == "" {
		return ErrForbidden
	}
	var visibility string
	var member bool
	err := r.db.QueryRowContext(ctx, `
		SELECT c.visibility, EXISTS(SELECT 1 FROM club_members m WHERE m.club_id = c.id AND m.user_id = $2)
		FROM clubs c WHERE c.id = $1`, clubID, userID).Scan(&visibility, &member)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("authorize club access: %w", err)
	}
	if !member {
		return ErrForbidden
	}
	return nil
}

func (r *PostgresRepository) requireModerator(ctx context.Context, clubID, userID string) error {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2`, clubID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("authorize club moderation: %w", err)
	}
	if role != "owner" && role != "admin" && role != "moderator" {
		return ErrForbidden
	}
	return nil
}

func (r *PostgresRepository) requireOwner(ctx context.Context, clubID, userID string) error {
	var role string
	err := r.db.QueryRowContext(ctx, `
		SELECT role FROM club_members WHERE club_id = $1 AND user_id = $2`, clubID, userID).Scan(&role)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrForbidden
	}
	if err != nil {
		return fmt.Errorf("authorize club owner: %w", err)
	}
	if role != "owner" {
		return ErrForbidden
	}
	return nil
}

func acceptedOrRejected(approve bool) string {
	if approve {
		return "accepted"
	}
	return "rejected"
}

var _ Repository = (*PostgresRepository)(nil)
