package auth

import (
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
)

type Repository interface {
	CreateUser(context.Context, RegisterInput, string) (User, error)
	FindByEmail(context.Context, string) (User, error)
	FindByID(context.Context, string) (User, error)
}

type SQLRepository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *SQLRepository {
	return &SQLRepository{db: db}
}

func (r *SQLRepository) CreateUser(ctx context.Context, input RegisterInput, passwordHash string) (User, error) {
	id, err := newUUID()
	if err != nil {
		return User{}, err
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return User{}, fmt.Errorf("begin registration transaction: %w", err)
	}
	defer tx.Rollback()

	var user User
	err = tx.QueryRowContext(ctx, `
		INSERT INTO users (id, email, password_hash)
		VALUES ($1, $2, $3)
		RETURNING id::text, email, created_at, updated_at`,
		id, input.Email, passwordHash,
	).Scan(&user.ID, &user.Email, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if isUniqueViolation(err, "users_email_key") {
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("insert user: %w", err)
	}
	err = tx.QueryRowContext(ctx, `
		INSERT INTO profiles (user_id, username, display_name)
		VALUES ($1, $2, $3)
		RETURNING username, display_name`,
		id, input.Username, input.DisplayName,
	).Scan(&user.Username, &user.DisplayName)
	if err != nil {
		if isUniqueViolation(err, "profiles_username_key") {
			return User{}, ErrUsernameTaken
		}
		return User{}, fmt.Errorf("insert user profile: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return User{}, fmt.Errorf("commit registration: %w", err)
	}
	return user, nil
}

func (r *SQLRepository) FindByEmail(ctx context.Context, email string) (User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `
		SELECT u.id::text, u.email, u.password_hash, p.display_name, p.username, u.created_at, u.updated_at
		FROM users u JOIN profiles p ON p.user_id = u.id
		WHERE u.email = $1 AND u.status = 'active'`, email))
}

func (r *SQLRepository) FindByID(ctx context.Context, id string) (User, error) {
	return scanUser(r.db.QueryRowContext(ctx, `
		SELECT u.id::text, u.email, u.password_hash, p.display_name, p.username, u.created_at, u.updated_at
		FROM users u JOIN profiles p ON p.user_id = u.id
		WHERE u.id = $1 AND u.status = 'active'`, id))
}

type rowScanner interface {
	Scan(...any) error
}

func scanUser(row rowScanner) (User, error) {
	var user User
	err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return User{}, ErrUserNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("read user: %w", err)
	}
	return user, nil
}

func isUniqueViolation(err error, constraint string) bool {
	var pgError *pgconn.PgError
	return errors.As(err, &pgError) && pgError.Code == "23505" && pgError.ConstraintName == constraint
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func newUUID() (string, error) {
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return "", fmt.Errorf("generate user ID: %w", err)
	}
	id[6] = (id[6] & 0x0f) | 0x40
	id[8] = (id[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16]), nil
}
