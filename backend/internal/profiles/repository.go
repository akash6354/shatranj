package profiles

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	FindByUsername(context.Context, string) (Profile, error)
	FindByUserID(context.Context, string) (Profile, error)
	Update(context.Context, Profile) (Profile, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) FindByUsername(ctx context.Context, username string) (Profile, error) {
	return scanProfile(r.db.QueryRowContext(ctx, profileSelect+` WHERE p.username = $1`, username))
}

func (r *SQLRepository) FindByUserID(ctx context.Context, userID string) (Profile, error) {
	return scanProfile(r.db.QueryRowContext(ctx, profileSelect+` WHERE p.user_id = $1`, userID))
}

func (r *SQLRepository) Update(ctx context.Context, profile Profile) (Profile, error) {
	result, err := r.db.ExecContext(ctx, `
		UPDATE profiles
		SET username = $2, display_name = $3, avatar_url = $4, country = $5, bio = $6, updated_at = now()
		WHERE user_id = $1`,
		profile.UserID, profile.Username, profile.DisplayName, nullable(profile.AvatarURL),
		nullable(profile.Country), nullable(profile.Bio),
	)
	if err != nil {
		var pgError *pgconn.PgError
		if errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == "profiles_username_key" {
			return Profile{}, ErrUsernameTaken
		}
		return Profile{}, fmt.Errorf("update profile: %w", err)
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return Profile{}, fmt.Errorf("check profile update: %w", err)
	}
	if rows == 0 {
		return Profile{}, ErrNotFound
	}
	return r.FindByUserID(ctx, profile.UserID)
}

const profileSelect = `
	SELECT p.user_id::text, p.username, p.display_name, COALESCE(p.avatar_url, ''),
	       COALESCE(p.country, ''), COALESCE(p.bio, ''),
	       p.ratings_bullet, p.ratings_blitz, p.ratings_rapid, p.ratings_classical,
	       p.created_at, p.updated_at
	FROM profiles p`

func scanProfile(row interface{ Scan(...any) error }) (Profile, error) {
	var profile Profile
	err := row.Scan(
		&profile.UserID, &profile.Username, &profile.DisplayName, &profile.AvatarURL,
		&profile.Country, &profile.Bio,
		&profile.Ratings.Bullet, &profile.Ratings.Blitz, &profile.Ratings.Rapid, &profile.Ratings.Classical,
		&profile.CreatedAt, &profile.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return Profile{}, ErrNotFound
	}
	if err != nil {
		return Profile{}, fmt.Errorf("read profile: %w", err)
	}
	return profile, nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
