package subscriptions

import (
	"context"
	"strings"
	"time"
)

type Service struct {
	repository   Repository
	premiumPrice int64
}

func NewService(repository Repository, premiumPricePaise int64) *Service {
	return &Service{repository: repository, premiumPrice: premiumPricePaise}
}

func (s *Service) Plans() []Plan {
	return []Plan{
		{ID: PlanFree, Name: "Free", Currency: "INR", Interval: "none", Entitlements: []string{}},
		{ID: PlanPremium, Name: "Premium", PricePaise: s.premiumPrice, Currency: "INR", Interval: "month",
			Entitlements: []string{EntitlementUnlimitedAnalysis, EntitlementAdvancedPuzzles}},
	}
}

func (s *Service) Get(ctx context.Context, userID string) (Subscription, error) {
	if strings.TrimSpace(userID) == "" {
		return Subscription{}, ErrInvalid
	}
	return s.repository.Get(ctx, userID)
}

func (s *Service) HasEntitlement(ctx context.Context, userID, entitlement string) (bool, error) {
	if strings.TrimSpace(userID) == "" || (entitlement != EntitlementUnlimitedAnalysis && entitlement != EntitlementAdvancedPuzzles) {
		return false, ErrInvalid
	}
	subscription, err := s.repository.Get(ctx, userID)
	if err != nil {
		return false, err
	}
	return subscription.PlanID == PlanPremium && subscription.Status == "active" &&
		subscription.EndsAt != nil && subscription.EndsAt.After(time.Now()), nil
}

func (s *Service) ActivatePremium(ctx context.Context, userID string, start, end time.Time) error {
	if strings.TrimSpace(userID) == "" || !end.After(start) {
		return ErrInvalid
	}
	return s.repository.ActivatePremium(ctx, userID, start, end)
}

func (s *Service) ActivateForPayment(ctx context.Context, userID string, start, end time.Time, paymentID string) error {
	if strings.TrimSpace(userID) == "" || strings.TrimSpace(paymentID) == "" || !end.After(start) {
		return ErrInvalid
	}
	return s.repository.ActivateForPayment(ctx, userID, start, end, paymentID)
}

func (s *Service) Cancel(ctx context.Context, userID string) error {
	if strings.TrimSpace(userID) == "" {
		return ErrInvalid
	}
	return s.repository.Cancel(ctx, userID)
}
