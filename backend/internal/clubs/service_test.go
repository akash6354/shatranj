package clubs

import (
	"context"
	"errors"
	"strings"
	"testing"
)

const testClubID = "11111111-1111-4111-8111-111111111111"
const testUserID = "22222222-2222-4222-8222-222222222222"

// stubRepository records the input passed to Create and satisfies Repository.
type stubRepository struct {
	created CreateInput
}

func (r *stubRepository) List(context.Context, string) ([]Club, error) { return nil, nil }
func (r *stubRepository) Create(_ context.Context, userID string, input CreateInput) (Club, error) {
	r.created = input
	return Club{ID: testClubID, OwnerID: userID, Name: input.Name, Visibility: input.Visibility}, nil
}
func (r *stubRepository) Get(context.Context, string, string) (Club, error) { return Club{}, nil }
func (r *stubRepository) Members(context.Context, string, string) ([]Member, error) {
	return nil, nil
}
func (r *stubRepository) Join(context.Context, string, string) (string, error) { return "", nil }
func (r *stubRepository) Leave(context.Context, string, string) error          { return nil }
func (r *stubRepository) Requests(context.Context, string, string) ([]JoinRequest, error) {
	return nil, nil
}
func (r *stubRepository) ResolveRequest(context.Context, string, string, string, bool) error {
	return nil
}
func (r *stubRepository) SetRole(context.Context, string, string, string, string) error { return nil }
func (r *stubRepository) Authorize(context.Context, string, string) error               { return nil }

func TestCreateValidatesInput(t *testing.T) {
	tests := []struct {
		name  string
		input CreateInput
	}{
		{name: "empty name", input: CreateInput{Name: "  "}},
		{name: "name too short", input: CreateInput{Name: "ab"}},
		{name: "name with illegal leading char", input: CreateInput{Name: "-bad"}},
		{name: "description too long", input: CreateInput{Name: "Valid Club", Description: strings.Repeat("x", 1001)}},
		{name: "unknown visibility", input: CreateInput{Name: "Valid Club", Visibility: "secret"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			service := NewService(new(stubRepository))
			if _, err := service.Create(context.Background(), testUserID, test.input); !errors.Is(err, ErrInvalid) {
				t.Fatalf("Create() error = %v, want ErrInvalid", err)
			}
		})
	}
}

func TestCreateDefaultsVisibilityToPublic(t *testing.T) {
	repository := new(stubRepository)
	service := NewService(repository)
	club, err := service.Create(context.Background(), testUserID, CreateInput{Name: "My Club"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if repository.created.Visibility != "public" || club.Visibility != "public" {
		t.Fatalf("visibility = %q, want public", club.Visibility)
	}
}

func TestCreateRejectsEmptyOwner(t *testing.T) {
	service := NewService(new(stubRepository))
	if _, err := service.Create(context.Background(), "", CreateInput{Name: "My Club"}); !errors.Is(err, ErrInvalid) {
		t.Fatalf("Create() error = %v, want ErrInvalid", err)
	}
}

func TestSetRoleValidatesRoleAndIdentifiers(t *testing.T) {
	service := NewService(new(stubRepository))
	ctx := context.Background()
	if err := service.SetRole(ctx, testClubID, testUserID, testUserID, "king"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("SetRole() with bad role error = %v, want ErrInvalid", err)
	}
	if err := service.SetRole(ctx, "not-a-uuid", testUserID, testUserID, "admin"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("SetRole() with bad club id error = %v, want ErrInvalid", err)
	}
	if err := service.SetRole(ctx, testClubID, testUserID, testUserID, "admin"); err != nil {
		t.Fatalf("SetRole() error = %v, want nil", err)
	}
}
