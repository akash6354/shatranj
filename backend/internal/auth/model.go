package auth

import (
	"errors"
	"time"
)

var (
	ErrInvalidInput       = errors.New("invalid authentication input")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailTaken         = errors.New("email is already registered")
	ErrUsernameTaken      = errors.New("username is already taken")
	ErrUserNotFound       = errors.New("user not found")
	ErrTokenInvalid       = errors.New("invalid or expired access token")
)

type User struct {
	ID           string
	Email        string
	PasswordHash string `json:"-"`
	DisplayName  string
	Username     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type RegisterInput struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken string     `json:"access_token"`
	TokenType   string     `json:"token_type"`
	ExpiresAt   time.Time  `json:"expires_at"`
	User        PublicUser `json:"user"`
}

type PublicUser struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Username    string `json:"username"`
}
