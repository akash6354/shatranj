package auth

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const minimumSigningKeyLength = 32

type TokenClaims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// TokenManager signs and verifies short-lived access tokens.
type TokenManager struct {
	key []byte
	ttl time.Duration
	now func() time.Time
}

func NewTokenManager(secret string, ttl time.Duration) (*TokenManager, error) {
	if len(secret) < minimumSigningKeyLength {
		return nil, fmt.Errorf("JWT signing secret must be at least %d bytes", minimumSigningKeyLength)
	}
	if ttl <= 0 {
		return nil, fmt.Errorf("JWT lifetime must be positive")
	}
	return &TokenManager{key: []byte(secret), ttl: ttl, now: time.Now}, nil
}

func (m *TokenManager) Issue(user User) (string, time.Time, error) {
	now := m.now().UTC()
	expires := now.Add(m.ttl)
	claims := TokenClaims{
		Email: user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   user.ID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expires),
			NotBefore: jwt.NewNumericDate(now),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.key)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign access token: %w", err)
	}
	return signed, expires, nil
}

func (m *TokenManager) Verify(raw string) (TokenClaims, error) {
	claims := new(TokenClaims)
	token, err := jwt.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, ErrTokenInvalid
		}
		return m.key, nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}), jwt.WithTimeFunc(m.now))
	if err != nil || token == nil || !token.Valid || claims.Subject == "" {
		return TokenClaims{}, ErrTokenInvalid
	}
	return *claims, nil
}
