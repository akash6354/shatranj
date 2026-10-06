package admin

import (
	"context"
	"strings"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) Dashboard(ctx context.Context) (Dashboard, error) {
	return s.repository.Dashboard(ctx)
}

func (s *Service) Users(ctx context.Context, query string, limit, offset int) ([]User, error) {
	query = strings.TrimSpace(query)
	if len(query) > 200 || offset < 0 {
		return nil, ErrInvalid
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repository.ListUsers(ctx, query, limit, offset)
}

func (s *Service) SetUserStatus(ctx context.Context, actorID, userID, status string) error {
	if !validID(userID) || (status != "active" && status != "suspended") {
		return ErrInvalid
	}
	if actorID == userID && status != "active" {
		return ErrForbidden
	}
	return s.repository.SetUserStatus(ctx, userID, status)
}

func (s *Service) Payments(ctx context.Context, limit, offset int) ([]Payment, error) {
	if offset < 0 {
		return nil, ErrInvalid
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repository.ListPayments(ctx, limit, offset)
}

func (s *Service) FlagPayment(ctx context.Context, adminID, paymentID, reason string) (PaymentFlag, error) {
	reason = strings.TrimSpace(reason)
	if !validID(adminID) || !validID(paymentID) || reason == "" || len(reason) > 1000 {
		return PaymentFlag{}, ErrInvalid
	}
	return s.repository.FlagPayment(ctx, paymentID, adminID, reason)
}

func (s *Service) PaymentFlags(ctx context.Context, limit, offset int) ([]PaymentFlag, error) {
	if offset < 0 {
		return nil, ErrInvalid
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.repository.ListPaymentFlags(ctx, limit, offset)
}

func (s *Service) ResolvePaymentFlag(ctx context.Context, adminID, flagID string) error {
	if !validID(adminID) || !validID(flagID) {
		return ErrInvalid
	}
	return s.repository.ResolvePaymentFlag(ctx, flagID, adminID)
}

func (s *Service) SetCoachStatus(ctx context.Context, userID, status string) error {
	if !validID(userID) || (status != "active" && status != "rejected" && status != "pending") {
		return ErrInvalid
	}
	return s.repository.SetCoachStatus(ctx, userID, status)
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
