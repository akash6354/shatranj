package subscriptions

import (
	"context"
	"testing"
	"time"
)

type repositoryStub struct{ subscription Subscription }

func (r repositoryStub) Get(context.Context, string) (Subscription, error) {
	return r.subscription, nil
}
func (repositoryStub) ActivatePremium(context.Context, string, time.Time, time.Time) error {
	return nil
}
func (repositoryStub) ActivateForPayment(context.Context, string, time.Time, time.Time, string) error {
	return nil
}
func (repositoryStub) Cancel(context.Context, string) error { return nil }

func TestHasEntitlementRequiresActiveUnexpiredPremium(t *testing.T) {
	now := time.Now()
	service := NewService(repositoryStub{subscription: Subscription{
		PlanID: PlanPremium, Status: "active", EndsAt: &now,
	}}, 49900)
	enabled, err := service.HasEntitlement(context.Background(), "user", EntitlementAdvancedPuzzles)
	if err != nil {
		t.Fatalf("HasEntitlement() error = %v", err)
	}
	if enabled {
		t.Fatal("expired subscription granted entitlement")
	}
}
