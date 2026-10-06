package coaches

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
)

type PostgresRepository struct{ db *sql.DB }

func NewRepository(db *sql.DB) *PostgresRepository { return &PostgresRepository{db: db} }

const coachSelect = `
	SELECT c.user_id::text, p.username, p.display_name, c.title, c.bio, c.rate_paise, c.currency,
		c.specialties, c.availability, c.status, c.created_at, c.updated_at
	FROM coach_profiles c JOIN profiles p ON p.user_id = c.user_id`

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]Coach, error) {
	rows, err := r.db.QueryContext(ctx, coachSelect+`
		WHERE c.status = 'active' ORDER BY c.rate_paise, p.username LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list coaches: %w", err)
	}
	defer rows.Close()
	items := make([]Coach, 0)
	for rows.Next() {
		item, err := scanCoach(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate coaches: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) Get(ctx context.Context, id string) (Coach, error) {
	item, err := scanCoach(r.db.QueryRowContext(ctx, coachSelect+` WHERE c.user_id = $1 AND c.status = 'active'`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Coach{}, ErrNotFound
	}
	return item, err
}

func (r *PostgresRepository) Upsert(ctx context.Context, userID string, input ProfileInput) (Coach, error) {
	specialties, err := json.Marshal(input.Specialties)
	if err != nil {
		return Coach{}, fmt.Errorf("encode coach specialties: %w", err)
	}
	availability, err := json.Marshal(input.Availability)
	if err != nil {
		return Coach{}, fmt.Errorf("encode coach availability: %w", err)
	}
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO coach_profiles (user_id, title, bio, rate_paise, specialties, availability)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (user_id) DO UPDATE SET title = EXCLUDED.title, bio = EXCLUDED.bio,
			rate_paise = EXCLUDED.rate_paise, specialties = EXCLUDED.specialties,
			availability = EXCLUDED.availability, updated_at = now()`,
		userID, input.Title, input.Bio, input.RatePaise, specialties, availability)
	if err != nil {
		return Coach{}, fmt.Errorf("save coach profile: %w", err)
	}
	return scanCoach(r.db.QueryRowContext(ctx, coachSelect+` WHERE c.user_id = $1`, userID))
}

func (r *PostgresRepository) CreateBooking(ctx context.Context, studentID, coachID string, input BookingInput) (Booking, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return Booking{}, fmt.Errorf("begin coach booking: %w", err)
	}
	defer tx.Rollback()
	var status string
	if err := tx.QueryRowContext(ctx, `
		SELECT status FROM coach_profiles WHERE user_id = $1 FOR UPDATE`, coachID).Scan(&status); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Booking{}, ErrNotFound
		}
		return Booking{}, fmt.Errorf("lock coach profile for booking: %w", err)
	}
	if status != "active" {
		return Booking{}, ErrNotFound
	}
	var busy bool
	if err := tx.QueryRowContext(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM coach_bookings b
			WHERE b.coach_id = $1 AND b.status IN ('pending', 'accepted')
				AND b.starts_at < $2 + ($3 * interval '1 minute')
				AND b.starts_at + (b.duration_minutes * interval '1 minute') > $2
		)`, coachID, input.StartsAt, input.DurationMin).Scan(&busy); err != nil {
		return Booking{}, fmt.Errorf("check coach booking availability: %w", err)
	}
	if busy {
		return Booking{}, ErrConflict
	}
	var booking Booking
	err = tx.QueryRowContext(ctx, `
		INSERT INTO coach_bookings (coach_id, student_id, starts_at, duration_minutes, notes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id::text, coach_id::text, student_id::text, starts_at, duration_minutes,
			notes, status, created_at, updated_at`,
		coachID, studentID, input.StartsAt, input.DurationMin, input.Notes).
		Scan(&booking.ID, &booking.CoachID, &booking.StudentID, &booking.StartsAt,
			&booking.DurationMin, &booking.Notes, &booking.Status, &booking.CreatedAt, &booking.UpdatedAt)
	if err != nil {
		return Booking{}, fmt.Errorf("create coach booking: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return Booking{}, fmt.Errorf("commit coach booking: %w", err)
	}
	return booking, nil
}

func (r *PostgresRepository) ListBookings(ctx context.Context, userID string) ([]Booking, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id::text, coach_id::text, student_id::text, starts_at, duration_minutes,
			notes, status, created_at, updated_at
		FROM coach_bookings WHERE coach_id = $1 OR student_id = $1
		ORDER BY starts_at DESC`, userID)
	if err != nil {
		return nil, fmt.Errorf("list coach bookings: %w", err)
	}
	defer rows.Close()
	items := make([]Booking, 0)
	for rows.Next() {
		var item Booking
		if err := rows.Scan(&item.ID, &item.CoachID, &item.StudentID, &item.StartsAt, &item.DurationMin,
			&item.Notes, &item.Status, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan coach booking: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate coach bookings: %w", err)
	}
	return items, nil
}

func (r *PostgresRepository) UpdateBookingStatus(ctx context.Context, userID, bookingID, status string) (Booking, error) {
	var item Booking
	err := r.db.QueryRowContext(ctx, `
		UPDATE coach_bookings b SET status = $3, updated_at = now()
		WHERE b.id = $1 AND (
			(b.coach_id = $2 AND ((b.status = 'pending' AND $3 IN ('accepted', 'declined', 'cancelled'))
				OR (b.status = 'accepted' AND $3 = 'cancelled')))
			OR (b.student_id = $2 AND b.status IN ('pending', 'accepted') AND $3 = 'cancelled')
		)
		RETURNING id::text, coach_id::text, student_id::text, starts_at, duration_minutes,
			notes, status, created_at, updated_at`, bookingID, userID, status).
		Scan(&item.ID, &item.CoachID, &item.StudentID, &item.StartsAt, &item.DurationMin,
			&item.Notes, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Booking{}, ErrNotFound
	}
	if err != nil {
		return Booking{}, fmt.Errorf("update coach booking: %w", err)
	}
	return item, nil
}

func (r *PostgresRepository) SetStatus(ctx context.Context, coachID, status string) error {
	result, err := r.db.ExecContext(ctx, `
		UPDATE coach_profiles SET status = $2, updated_at = now() WHERE user_id = $1`, coachID, status)
	if err != nil {
		return fmt.Errorf("moderate coach profile: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read coach moderation result: %w", err)
	}
	if count == 0 {
		return ErrNotFound
	}
	return nil
}

type scanner interface{ Scan(...any) error }

func scanCoach(row scanner) (Coach, error) {
	var item Coach
	var specialties, availability []byte
	err := row.Scan(&item.UserID, &item.Username, &item.DisplayName, &item.Title, &item.Bio,
		&item.RatePaise, &item.Currency, &specialties, &availability, &item.Status, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return Coach{}, fmt.Errorf("read coach profile: %w", err)
	}
	if err := json.Unmarshal(specialties, &item.Specialties); err != nil {
		return Coach{}, fmt.Errorf("decode coach specialties: %w", err)
	}
	if len(availability) > 0 {
		if err := json.Unmarshal(availability, &item.Availability); err != nil {
			return Coach{}, fmt.Errorf("decode coach availability: %w", err)
		}
	}
	return item, nil
}

var _ Repository = (*PostgresRepository)(nil)
