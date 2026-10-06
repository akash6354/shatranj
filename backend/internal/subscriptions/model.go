package subscriptions

import (
	"context"
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("subscription not found")
	ErrInvalid  = errors.New("invalid subscription request")
)

type Plan struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	PricePaise   int64    `json:"price_paise"`
	Currency     string   `json:"currency"`
	Interval     string   `json:"interval"`
	Entitlements []string `json:"entitlements"`
}

type Subscription struct {
	UserID    string     `json:"user_id"`
	PlanID    string     `json:"plan_id"`
	Status    string     `json:"status"`
	StartedAt *time.Time `json:"started_at,omitempty"`
	EndsAt    *time.Time `json:"ends_at,omitempty"`
	UpdatedAt time.Time  `json:"updated_at,omitempty"`
}

const (
	PlanFree                     = "free"
	PlanPremium                  = "premium"
	EntitlementUnlimitedAnalysis = "unlimited_analysis"
	EntitlementAdvancedPuzzles   = "advanced_puzzles"
)

type Repository interface {
	Get(context.Context, string) (Subscription, error)
	ActivatePremium(context.Context, string, time.Time, time.Time) error
	ActivateForPayment(context.Context, string, time.Time, time.Time, string) error
	Cancel(context.Context, string) error
}
