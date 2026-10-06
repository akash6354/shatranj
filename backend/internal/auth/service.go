package auth

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_]{3,24}$`)

type Service struct {
	repository Repository
	tokens     *TokenManager
}

func NewService(repository Repository, tokens *TokenManager) *Service {
	return &Service{repository: repository, tokens: tokens}
}

func (s *Service) Register(ctx context.Context, input RegisterInput) (AuthResponse, error) {
	input.Email = normalizeEmail(input.Email)
	input.Username = strings.ToLower(strings.TrimSpace(input.Username))
	input.DisplayName = strings.TrimSpace(input.DisplayName)
	parsedEmail, err := mail.ParseAddress(input.Email)
	if err != nil || parsedEmail.Address != input.Email {
		return AuthResponse{}, fmt.Errorf("%w: valid email is required", ErrInvalidInput)
	}
	if len(input.Email) > 254 || strings.ContainsAny(input.Email, " \t\r\n") {
		return AuthResponse{}, fmt.Errorf("%w: valid email is required", ErrInvalidInput)
	}
	if len(input.Password) < 12 || len(input.Password) > 72 {
		return AuthResponse{}, fmt.Errorf("%w: password must be between 12 and 72 bytes", ErrInvalidInput)
	}
	if !usernamePattern.MatchString(input.Username) {
		return AuthResponse{}, fmt.Errorf("%w: username must be 3-24 letters, numbers, or underscores", ErrInvalidInput)
	}
	if len([]rune(input.DisplayName)) < 1 || len([]rune(input.DisplayName)) > 80 {
		return AuthResponse{}, fmt.Errorf("%w: display name must be 1-80 characters", ErrInvalidInput)
	}
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, fmt.Errorf("hash password: %w", err)
	}
	user, err := s.repository.CreateUser(ctx, input, string(passwordHash))
	if err != nil {
		return AuthResponse{}, err
	}
	return s.createResponse(user)
}

func (s *Service) Login(ctx context.Context, input LoginInput) (AuthResponse, error) {
	email := normalizeEmail(input.Email)
	if email == "" || input.Password == "" {
		return AuthResponse{}, ErrInvalidCredentials
	}
	user, err := s.repository.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return AuthResponse{}, ErrInvalidCredentials
		}
		return AuthResponse{}, err
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}
	return s.createResponse(user)
}

func (s *Service) UserByID(ctx context.Context, id string) (PublicUser, error) {
	user, err := s.repository.FindByID(ctx, id)
	if err != nil {
		return PublicUser{}, err
	}
	return publicUser(user), nil
}

func (s *Service) createResponse(user User) (AuthResponse, error) {
	token, expiresAt, err := s.tokens.Issue(user)
	if err != nil {
		return AuthResponse{}, err
	}
	return AuthResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresAt:   expiresAt,
		User:        publicUser(user),
	}, nil
}

func publicUser(user User) PublicUser {
	return PublicUser{ID: user.ID, Email: user.Email, DisplayName: user.DisplayName, Username: user.Username}
}
