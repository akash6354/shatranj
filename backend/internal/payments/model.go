package payments

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalid             = errors.New("invalid payment request")
	ErrNotFound            = errors.New("payment not found")
	ErrProviderUnavailable = errors.New("payment provider unavailable")
	ErrSignatureInvalid    = errors.New("invalid payment webhook signature")
)

type Payment struct {
	ID                string    `json:"id"`
	UserID            string    `json:"user_id,omitempty"`
	Provider          string    `json:"provider"`
	ProviderOrderID   string    `json:"provider_order_id,omitempty"`
	ProviderPaymentID string    `json:"provider_payment_id,omitempty"`
	AmountPaise       int64     `json:"amount_paise"`
	Currency          string    `json:"currency"`
	Status            string    `json:"status"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type Repository interface {
	CreatePending(context.Context, string, int64, string) (Payment, error)
	SetProviderOrder(context.Context, string, string) (Payment, error)
	MarkFailed(context.Context, string) error
	Capture(context.Context, string, string, int64, string) (Payment, error)
	List(context.Context, string, int, int) ([]Payment, error)
}

type Provider interface {
	CreateOrder(context.Context, int64, string, string) (string, error)
	KeyID() string
	VerifyWebhook([]byte, string) bool
}

type SubscriptionActivator interface {
	ActivateForPayment(context.Context, string, time.Time, time.Time, string) error
}
