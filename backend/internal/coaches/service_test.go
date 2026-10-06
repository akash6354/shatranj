package coaches

import (
	"context"
	"errors"
	"testing"
)

func TestUpsertRejectsInvalidProfileBeforeRepositoryAccess(t *testing.T) {
	service := NewService(nil)
	_, err := service.Upsert(context.Background(), "00000000-0000-4000-8000-000000000001", ProfileInput{})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Upsert() error = %v, want ErrInvalid", err)
	}
}
