package notifications

import (
	"context"
	"errors"
	"testing"
)

type repositoryStub struct{ created Notification }

func (r *repositoryStub) Create(_ context.Context, input CreateInput) (Notification, error) {
	r.created = Notification{UserID: input.UserID, Type: input.Type, Title: input.Title, Body: input.Body, Data: input.Data}
	return r.created, nil
}
func (*repositoryStub) List(context.Context, string, bool, int, int) ([]Notification, error) {
	return nil, nil
}
func (*repositoryStub) UnreadCount(context.Context, string) (int, error) { return 0, nil }
func (*repositoryStub) SetRead(context.Context, string, string, bool) (Notification, error) {
	return Notification{}, nil
}

func TestCreatePersistsNotificationAndSurfacesPushFailure(t *testing.T) {
	repository := &repositoryStub{}
	pushErr := errors.New("push provider unavailable")
	service := NewService(repository, PushFunc(func(context.Context, Notification) error { return pushErr }))
	item, err := service.Create(context.Background(), CreateInput{
		UserID: "00000000-0000-4000-8000-000000000001",
		Type:   "game.ended",
		Title:  " Game finished ",
		Data:   map[string]string{"game_id": "game-id"},
	})
	if !errors.Is(err, pushErr) {
		t.Fatalf("Create() error = %v, want push error", err)
	}
	if item.UserID != repository.created.UserID || item.Title != "Game finished" {
		t.Fatalf("notification = %+v; repository created %+v", item, repository.created)
	}
}

func TestCreateRejectsInvalidUserBeforePersistence(t *testing.T) {
	service := NewService(&repositoryStub{}, nil)
	_, err := service.Create(context.Background(), CreateInput{UserID: "invalid", Type: "test", Title: "title"})
	if !errors.Is(err, ErrInvalid) {
		t.Fatalf("Create() error = %v, want ErrInvalid", err)
	}
}
