package chat

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

func (r *PostgresRepository) ListRooms(ctx context.Context, userID string) ([]Room, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT cr.id::text, cr.kind, COALESCE(cr.reference_id::text, ''), cr.created_at
		FROM chat_rooms cr
		LEFT JOIN chat_room_members rm ON rm.room_id = cr.id AND rm.user_id = $1
		LEFT JOIN games g ON cr.kind = 'game' AND g.id = cr.reference_id
		LEFT JOIN club_members cm ON cr.kind = 'club' AND cm.club_id = cr.reference_id AND cm.user_id = $1
		WHERE (cr.kind = 'direct' AND rm.user_id IS NOT NULL AND EXISTS (
				SELECT 1 FROM friendships f
				WHERE f.user_id < f.friend_id AND f.status = 'accepted'
				  AND cr.direct_key = f.user_id::text || ':' || f.friend_id::text
			))
			OR (cr.kind = 'game' AND (g.white_player_id = $1 OR g.black_player_id = $1))
			OR cm.user_id IS NOT NULL
		ORDER BY cr.created_at DESC, cr.id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list chat rooms: %w", err)
	}
	defer rows.Close()
	rooms := make([]Room, 0)
	for rows.Next() {
		var room Room
		if err := rows.Scan(&room.ID, &room.Kind, &room.Reference, &room.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat room: %w", err)
		}
		rooms = append(rooms, room)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat rooms: %w", err)
	}
	return rooms, nil
}

func (r *PostgresRepository) Open(ctx context.Context, userID string, input CreateRoomInput) (Room, error) {
	switch input.Kind {
	case GameRoom:
		return r.openGame(ctx, userID, input.ReferenceID)
	case ClubRoom:
		return r.openClub(ctx, userID, input.ReferenceID)
	case DirectRoom:
		return r.openDirect(ctx, userID, input.UserID)
	default:
		return Room{}, ErrInvalid
	}
}

func (r *PostgresRepository) openGame(ctx context.Context, userID, gameID string) (Room, error) {
	var room Room
	err := r.db.QueryRowContext(ctx, `
		WITH authorized AS (
			SELECT id FROM games WHERE id = $1 AND (white_player_id = $2 OR black_player_id = $2)
		), room AS (
			INSERT INTO chat_rooms (kind, reference_id)
			SELECT 'game', id FROM authorized
			ON CONFLICT (kind, reference_id) WHERE reference_id IS NOT NULL
			DO UPDATE SET updated_at = now()
			RETURNING id, kind, reference_id, created_at
		)
		SELECT id::text, kind, reference_id::text, created_at FROM room`, gameID, userID).Scan(
		&room.ID, &room.Kind, &room.Reference, &room.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Room{}, ErrForbidden
	}
	if err != nil {
		return Room{}, fmt.Errorf("open game chat room: %w", err)
	}
	return room, nil
}

func (r *PostgresRepository) openClub(ctx context.Context, userID, clubID string) (Room, error) {
	var room Room
	err := r.db.QueryRowContext(ctx, `
		WITH authorized AS (
			SELECT id FROM clubs c WHERE c.id = $1 AND EXISTS (
				SELECT 1 FROM club_members m WHERE m.club_id = c.id AND m.user_id = $2
			)
		), room AS (
			INSERT INTO chat_rooms (kind, reference_id)
			SELECT 'club', id FROM authorized
			ON CONFLICT (kind, reference_id) WHERE reference_id IS NOT NULL
			DO UPDATE SET updated_at = now()
			RETURNING id, kind, reference_id, created_at
		)
		SELECT id::text, kind, reference_id::text, created_at FROM room`, clubID, userID).Scan(
		&room.ID, &room.Kind, &room.Reference, &room.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Room{}, ErrForbidden
	}
	if err != nil {
		return Room{}, fmt.Errorf("open club chat room: %w", err)
	}
	return room, nil
}

func (r *PostgresRepository) openDirect(ctx context.Context, userID, targetID string) (Room, error) {
	a, b := strings.ToLower(userID), strings.ToLower(targetID)
	if a > b {
		a, b = b, a
	}
	directKey := a + ":" + b
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Room{}, fmt.Errorf("begin direct chat creation: %w", err)
	}
	defer tx.Rollback()
	var allowed bool
	err = tx.QueryRowContext(ctx, `
		SELECT EXISTS(SELECT 1 FROM friendships WHERE user_id = $1 AND friend_id = $2 AND status = 'accepted')`,
		a, b).Scan(&allowed)
	if err != nil {
		return Room{}, fmt.Errorf("check direct chat friendship: %w", err)
	}
	if !allowed {
		return Room{}, ErrForbidden
	}
	var room Room
	err = tx.QueryRowContext(ctx, `
		INSERT INTO chat_rooms (kind, direct_key) VALUES ('direct', $1)
		ON CONFLICT (direct_key) DO UPDATE SET updated_at = now()
		RETURNING id::text, kind, created_at`, directKey).Scan(&room.ID, &room.Kind, &room.CreatedAt)
	if err != nil {
		return Room{}, fmt.Errorf("create direct chat room: %w", err)
	}
	room.Reference = ""
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO chat_room_members (room_id, user_id) VALUES ($1, $2), ($1, $3)
		ON CONFLICT (room_id, user_id) DO NOTHING`, room.ID, a, b); err != nil {
		return Room{}, fmt.Errorf("add direct chat participants: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Room{}, fmt.Errorf("commit direct chat: %w", err)
	}
	return room, nil
}

func (r *PostgresRepository) Authorize(ctx context.Context, roomID, userID string) error {
	var allowed bool
	err := r.db.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM chat_rooms cr
			WHERE cr.id = $1 AND (
				(cr.kind = 'direct' AND EXISTS(
					SELECT 1 FROM chat_room_members rm JOIN friendships f
						ON cr.direct_key = f.user_id::text || ':' || f.friend_id::text
					WHERE rm.room_id = cr.id AND rm.user_id = $2 AND f.status = 'accepted'
				))
				OR (cr.kind IN ('game', 'club') AND EXISTS(
					SELECT 1 FROM chat_room_members rm WHERE rm.room_id = cr.id AND rm.user_id = $2
				))
				OR (cr.kind = 'game' AND EXISTS(
					SELECT 1 FROM games g WHERE g.id = cr.reference_id AND (g.white_player_id = $2 OR g.black_player_id = $2)
				))
				OR (cr.kind = 'club' AND EXISTS(
					SELECT 1 FROM club_members cm WHERE cm.club_id = cr.reference_id AND cm.user_id = $2
				))
			)
		)`, roomID, userID).Scan(&allowed)
	if err != nil {
		return fmt.Errorf("authorize chat room: %w", err)
	}
	if !allowed {
		var exists bool
		if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM chat_rooms WHERE id = $1)`, roomID).Scan(&exists); err != nil {
			return fmt.Errorf("check chat room: %w", err)
		}
		if !exists {
			return ErrNotFound
		}
		return ErrForbidden
	}
	return nil
}

func (r *PostgresRepository) Messages(ctx context.Context, roomID, _ string, limit int) ([]Message, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id::text, m.room_id::text, m.sender_id::text, COALESCE(p.username, ''),
			m.content, m.created_at
		FROM chat_messages m LEFT JOIN profiles p ON p.user_id = m.sender_id
		WHERE m.room_id = $1 AND m.status = 'visible'
		ORDER BY m.created_at DESC, m.id DESC LIMIT $2`, roomID, limit)
	if err != nil {
		return nil, fmt.Errorf("list chat messages: %w", err)
	}
	defer rows.Close()
	messages := make([]Message, 0)
	for rows.Next() {
		var message Message
		if err := rows.Scan(&message.ID, &message.RoomID, &message.SenderID,
			&message.Username, &message.Content, &message.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat message: %w", err)
		}
		messages = append(messages, message)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat messages: %w", err)
	}
	for left, right := 0, len(messages)-1; left < right; left, right = left+1, right-1 {
		messages[left], messages[right] = messages[right], messages[left]
	}
	return messages, nil
}

func (r *PostgresRepository) Send(ctx context.Context, roomID, userID, content string) (Message, error) {
	var message Message
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO chat_messages (room_id, sender_id, content)
		SELECT $1, $2, $3
		WHERE EXISTS(
			SELECT 1 FROM chat_rooms cr WHERE cr.id = $1 AND (
				(cr.kind = 'direct' AND EXISTS(
					SELECT 1 FROM chat_room_members rm JOIN friendships f
						ON cr.direct_key = f.user_id::text || ':' || f.friend_id::text
					WHERE rm.room_id = cr.id AND rm.user_id = $2 AND f.status = 'accepted'
				))
				OR (cr.kind IN ('game', 'club') AND EXISTS(
					SELECT 1 FROM chat_room_members rm WHERE rm.room_id = cr.id AND rm.user_id = $2
				))
				OR (cr.kind = 'game' AND EXISTS(
					SELECT 1 FROM games g WHERE g.id = cr.reference_id AND (g.white_player_id = $2 OR g.black_player_id = $2)
				))
				OR (cr.kind = 'club' AND EXISTS(
					SELECT 1 FROM club_members cm WHERE cm.club_id = cr.reference_id AND cm.user_id = $2
				))
			)
		)
		RETURNING id::text, room_id::text, sender_id::text, content, created_at`,
		roomID, userID, content).Scan(&message.ID, &message.RoomID, &message.SenderID, &message.Content, &message.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Message{}, ErrForbidden
	}
	if err != nil {
		return Message{}, fmt.Errorf("send chat message: %w", err)
	}
	if err := r.db.QueryRowContext(ctx, `SELECT COALESCE(username, '') FROM profiles WHERE user_id = $1`, userID).Scan(&message.Username); err != nil && !errors.Is(err, sql.ErrNoRows) {
		return Message{}, fmt.Errorf("load chat sender profile: %w", err)
	}
	return message, nil
}

func (r *PostgresRepository) Hide(ctx context.Context, roomID, messageID, moderatorID, reason string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE chat_messages m SET status = 'hidden', moderation_reason = $4
		FROM chat_rooms cr
		WHERE m.id = $2 AND m.room_id = $1 AND cr.id = m.room_id
		  AND (m.sender_id = $3 OR (cr.kind = 'club' AND EXISTS(
			SELECT 1 FROM club_members cm WHERE cm.club_id = cr.reference_id
				AND cm.user_id = $3 AND cm.role IN ('owner', 'admin', 'moderator')
		  )))`, roomID, messageID, moderatorID, reason)
	if err != nil {
		return fmt.Errorf("hide chat message: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read moderation result: %w", err)
	}
	if affected == 0 {
		return ErrForbidden
	}
	return nil
}

var _ Repository = (*PostgresRepository)(nil)
