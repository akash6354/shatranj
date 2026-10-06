package profiles

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotFound      = errors.New("profile not found")
	ErrUsernameTaken = errors.New("username is already taken")
	ErrInvalidInput  = errors.New("invalid profile input")
	profileUsername  = regexp.MustCompile(`^[a-zA-Z0-9_]{3,24}$`)
	profileCountry   = regexp.MustCompile(`^[A-Z]{2}$`)
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) GetByUsername(ctx context.Context, username string) (Profile, error) {
	return s.repository.FindByUsername(ctx, strings.ToLower(strings.TrimSpace(username)))
}

func (s *Service) GetByUserID(ctx context.Context, userID string) (Profile, error) {
	return s.repository.FindByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID string, input UpdateInput) (Profile, error) {
	profile, err := s.repository.FindByUserID(ctx, userID)
	if err != nil {
		return Profile{}, err
	}
	if input.Username != nil {
		profile.Username = strings.ToLower(strings.TrimSpace(*input.Username))
	}
	if input.DisplayName != nil {
		profile.DisplayName = strings.TrimSpace(*input.DisplayName)
	}
	if input.AvatarURL != nil {
		profile.AvatarURL = strings.TrimSpace(*input.AvatarURL)
	}
	if input.Country != nil {
		profile.Country = strings.ToUpper(strings.TrimSpace(*input.Country))
	}
	if input.Bio != nil {
		profile.Bio = strings.TrimSpace(*input.Bio)
	}

	if !profileUsername.MatchString(profile.Username) {
		return Profile{}, fmt.Errorf("%w: username must be 3-24 letters, numbers, or underscores", ErrInvalidInput)
	}
	if utf8.RuneCountInString(profile.DisplayName) < 1 || utf8.RuneCountInString(profile.DisplayName) > 80 {
		return Profile{}, fmt.Errorf("%w: display name must be 1-80 characters", ErrInvalidInput)
	}
	if profile.Country != "" && !profileCountry.MatchString(profile.Country) {
		return Profile{}, fmt.Errorf("%w: country must be a two-letter code", ErrInvalidInput)
	}
	if utf8.RuneCountInString(profile.Bio) > 500 {
		return Profile{}, fmt.Errorf("%w: bio must be at most 500 characters", ErrInvalidInput)
	}
	if profile.AvatarURL != "" {
		parsed, err := url.ParseRequestURI(profile.AvatarURL)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" {
			return Profile{}, fmt.Errorf("%w: avatar_url must be an absolute HTTP(S) URL", ErrInvalidInput)
		}
	}
	return s.repository.Update(ctx, profile)
}
