package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type memoryRepository struct {
	user User
}

func (r *memoryRepository) CreateUser(_ context.Context, input RegisterInput, passwordHash string) (User, error) {
	r.user = User{
		ID: "user-id", Email: input.Email, PasswordHash: passwordHash,
		DisplayName: input.DisplayName, Username: input.Username,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	return r.user, nil
}

func (r *memoryRepository) FindByEmail(_ context.Context, email string) (User, error) {
	if email != r.user.Email {
		return User{}, ErrUserNotFound
	}
	return r.user, nil
}

func (r *memoryRepository) FindByID(_ context.Context, id string) (User, error) {
	if id != r.user.ID {
		return User{}, ErrUserNotFound
	}
	return r.user, nil
}

func newTestService(t *testing.T) (*Service, *memoryRepository, *TokenManager) {
	t.Helper()
	tokens, err := NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	repository := new(memoryRepository)
	return NewService(repository, tokens), repository, tokens
}

func TestRegisterAndLogin(t *testing.T) {
	service, repository, tokens := newTestService(t)
	result, err := service.Register(context.Background(), RegisterInput{
		Email: "Player@Example.com", Password: "a long secure password", DisplayName: "Player", Username: "Chess_Player",
	})
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}
	if result.User.Email != "player@example.com" || result.User.Username != "chess_player" {
		t.Fatalf("registration user = %+v", result.User)
	}
	if repository.user.PasswordHash == "a long secure password" {
		t.Fatal("password was stored without hashing")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(repository.user.PasswordHash), []byte("a long secure password")); err != nil {
		t.Fatalf("stored password hash did not verify: %v", err)
	}
	claims, err := tokens.Verify(result.AccessToken)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.Subject != "user-id" || claims.Email != "player@example.com" {
		t.Fatalf("claims = %+v", claims)
	}
	login, err := service.Login(context.Background(), LoginInput{Email: "PLAYER@example.com", Password: "a long secure password"})
	if err != nil || login.User.ID != result.User.ID {
		t.Fatalf("Login() = %+v, %v", login, err)
	}
}

func TestLoginUsesGenericInvalidCredentials(t *testing.T) {
	service, _, _ := newTestService(t)
	_, err := service.Login(context.Background(), LoginInput{Email: "missing@example.com", Password: "incorrect"})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestTokenExpirationAndKeyValidation(t *testing.T) {
	now := time.Now().UTC()
	manager, err := NewTokenManager("0123456789abcdef0123456789abcdef", time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	manager.now = func() time.Time { return now }
	raw, _, err := manager.Issue(User{ID: "u1", Email: "user@example.com"})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := manager.Verify(raw); err != nil {
		t.Fatalf("valid token rejected: %v", err)
	}
	manager.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, err := manager.Verify(raw); !errors.Is(err, ErrTokenInvalid) {
		t.Fatalf("expired token error = %v, want ErrTokenInvalid", err)
	}
	if _, err := NewTokenManager("too-short", time.Minute); err == nil {
		t.Fatal("short JWT secret accepted")
	}
}

func TestOTPIsSixDigits(t *testing.T) {
	code, err := GenerateOTP()
	if err != nil {
		t.Fatal(err)
	}
	if len(code) != 6 {
		t.Fatalf("OTP %q has length %d, want 6", code, len(code))
	}
}
