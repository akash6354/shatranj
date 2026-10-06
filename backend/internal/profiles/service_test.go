package profiles

import (
	"context"
	"testing"
	"time"
)

type memoryProfiles struct {
	profile Profile
}

func (m *memoryProfiles) FindByUsername(_ context.Context, username string) (Profile, error) {
	if username != m.profile.Username {
		return Profile{}, ErrNotFound
	}
	return m.profile, nil
}

func (m *memoryProfiles) FindByUserID(_ context.Context, id string) (Profile, error) {
	if id != m.profile.UserID {
		return Profile{}, ErrNotFound
	}
	return m.profile, nil
}

func (m *memoryProfiles) Update(_ context.Context, profile Profile) (Profile, error) {
	profile.UpdatedAt = time.Now()
	m.profile = profile
	return profile, nil
}

func TestUpdateProfileFields(t *testing.T) {
	repository := &memoryProfiles{profile: Profile{
		UserID: "user-id", Username: "player", DisplayName: "Player",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}}
	service := NewService(repository)
	username, displayName, avatarURL, country, bio := "New_Player", "New Name", "https://example.com/avatar.png", "us", "Hello!"
	updated, err := service.Update(context.Background(), "user-id", UpdateInput{
		Username: &username, DisplayName: &displayName, AvatarURL: &avatarURL, Country: &country, Bio: &bio,
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	if updated.Username != "new_player" || updated.DisplayName != "New Name" || updated.Country != "US" {
		t.Fatalf("updated profile = %+v", updated)
	}
}

func TestUpdateRejectsInvalidValues(t *testing.T) {
	repository := &memoryProfiles{profile: Profile{UserID: "user-id", Username: "player", DisplayName: "Player"}}
	service := NewService(repository)
	badCountry := "USA"
	if _, err := service.Update(context.Background(), "user-id", UpdateInput{Country: &badCountry}); err == nil {
		t.Fatal("accepted an invalid country code")
	}
	badAvatar := "javascript:alert(1)"
	if _, err := service.Update(context.Background(), "user-id", UpdateInput{AvatarURL: &badAvatar}); err == nil {
		t.Fatal("accepted a non-HTTP avatar URL")
	}
}
