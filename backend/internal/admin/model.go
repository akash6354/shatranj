package admin

import (
	"context"
	"errors"
	"time"
)

var (
	ErrForbidden = errors.New("admin access required")
	ErrInvalid   = errors.New("invalid admin request")
	ErrNotFound  = errors.New("admin resource not found")
)

type Dashboard struct {
	ActiveUsers      int64 `json:"active_users"`
	SuspendedUsers   int64 `json:"suspended_users"`
	PendingPayments  int64 `json:"pending_payments"`
	PaidPayments     int64 `json:"paid_payments"`
	DraftNews        int64 `json:"draft_news"`
	PendingCoachApps int64 `json:"pending_coach_applications"`
	PendingBookings  int64 `json:"pending_coach_bookings"`
	OpenPaymentFlags int64 `json:"open_payment_flags"`
}

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Payment struct {
	ID              string    `json:"id"`
	UserID          string    `json:"user_id"`
	Username        string    `json:"username"`
	AmountPaise     int64     `json:"amount_paise"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	ProviderOrderID string    `json:"provider_order_id,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type PaymentFlag struct {
	ID        string    `json:"id"`
	PaymentID string    `json:"payment_id"`
	AdminID   string    `json:"admin_id"`
	Reason    string    `json:"reason"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Repository interface {
	IsAdmin(context.Context, string) (bool, error)
	Dashboard(context.Context) (Dashboard, error)
	ListUsers(context.Context, string, int, int) ([]User, error)
	SetUserStatus(context.Context, string, string) error
	ListPayments(context.Context, int, int) ([]Payment, error)
	FlagPayment(context.Context, string, string, string) (PaymentFlag, error)
	ListPaymentFlags(context.Context, int, int) ([]PaymentFlag, error)
	ResolvePaymentFlag(context.Context, string, string) error
	SetCoachStatus(context.Context, string, string) error
}
