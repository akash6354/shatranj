package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repository    Repository
	provider      Provider
	subscriptions SubscriptionActivator
	premiumPrice  int64
}

func NewService(repository Repository, provider Provider, subscriptions SubscriptionActivator, premiumPricePaise int64) *Service {
	return &Service{repository: repository, provider: provider, subscriptions: subscriptions, premiumPrice: premiumPricePaise}
}

func (s *Service) CreatePremiumOrder(ctx context.Context, userID string) (Payment, string, error) {
	if !validID(userID) || s.premiumPrice <= 0 {
		return Payment{}, "", ErrInvalid
	}
	if s.provider == nil || s.provider.KeyID() == "" {
		return Payment{}, "", ErrProviderUnavailable
	}
	payment, err := s.repository.CreatePending(ctx, userID, s.premiumPrice, "INR")
	if err != nil {
		return Payment{}, "", err
	}
	orderID, err := s.provider.CreateOrder(ctx, payment.AmountPaise, payment.Currency, payment.ID)
	if err != nil {
		if markErr := s.repository.MarkFailed(ctx, payment.ID); markErr != nil {
			return Payment{}, "", fmt.Errorf("create provider order (%v) and mark local order failed: %w", err, markErr)
		}
		return Payment{}, "", err
	}
	payment, err = s.repository.SetProviderOrder(ctx, payment.ID, orderID)
	if err != nil {
		if markErr := s.repository.MarkFailed(ctx, payment.ID); markErr != nil {
			return Payment{}, "", fmt.Errorf("save provider order (%v) and mark local order failed: %w", err, markErr)
		}
		return Payment{}, "", fmt.Errorf("save Razorpay order ID: %w", err)
	}
	return payment, s.provider.KeyID(), nil
}

func (s *Service) List(ctx context.Context, userID string, limit, offset int) ([]Payment, error) {
	if !validID(userID) {
		return nil, ErrInvalid
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		return nil, ErrInvalid
	}
	return s.repository.List(ctx, userID, limit, offset)
}

func (s *Service) HandleWebhook(ctx context.Context, body []byte, signature string) error {
	if len(body) == 0 || len(body) > 1<<20 {
		return ErrInvalid
	}
	if s.provider == nil {
		return ErrProviderUnavailable
	}
	if !s.provider.VerifyWebhook(body, signature) {
		return ErrSignatureInvalid
	}
	var event struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID       string `json:"id"`
					OrderID  string `json:"order_id"`
					Amount   int64  `json:"amount"`
					Currency string `json:"currency"`
				} `json:"entity"`
			} `json:"payment"`
			Order struct {
				Entity struct {
					ID       string `json:"id"`
					Amount   int64  `json:"amount"`
					Currency string `json:"currency"`
				} `json:"entity"`
			} `json:"order"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return ErrInvalid
	}
	if event.Event != "payment.captured" && event.Event != "order.paid" {
		return nil
	}
	p := event.Payload.Payment.Entity
	o := event.Payload.Order.Entity
	orderID, paymentID, amount, currency := p.OrderID, p.ID, p.Amount, p.Currency
	if orderID == "" {
		orderID, amount, currency = o.ID, o.Amount, o.Currency
	}
	if !strings.HasPrefix(orderID, "order_") || amount <= 0 || currency != "INR" {
		return ErrInvalid
	}
	payment, err := s.repository.Capture(ctx, orderID, paymentID, amount, currency)
	if err != nil {
		return err
	}
	if s.subscriptions == nil {
		return errors.New("subscription activation is not configured")
	}
	now := time.Now().UTC()
	return s.subscriptions.ActivateForPayment(ctx, payment.UserID, now, now.AddDate(0, 1, 0), payment.ID)
}

func validID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for index, value := range id {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			if value != '-' {
				return false
			}
		} else if !(value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F') {
			return false
		}
	}
	return true
}
