package auth

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
)

// OTPDelivery is the future email/SMS delivery seam. No delivery provider is
// configured in this stage.
type OTPDelivery interface {
	SendLoginCode(context.Context, string, string) error
}

// GenerateOTP creates an unpredictable six-digit login code. Callers must
// store only a short-lived hash and deliver it through a trusted provider.
func GenerateOTP() (string, error) {
	limit := big.NewInt(1_000_000)
	value, err := rand.Int(rand.Reader, limit)
	if err != nil {
		return "", fmt.Errorf("generate login code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}
