package news

import (
	"context"
	"strings"
	"unicode"
)

type Service struct{ repository Repository }

func NewService(repository Repository) *Service { return &Service{repository: repository} }

func (s *Service) List(ctx context.Context, limit, offset int) ([]Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	if offset < 0 {
		return nil, ErrInvalid
	}
	return s.repository.List(ctx, limit, offset)
}

func (s *Service) ListAdmin(ctx context.Context, status string, limit, offset int) ([]Post, error) {
	if status != "" && status != "draft" && status != "published" && status != "archived" {
		return nil, ErrInvalid
	}
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if offset < 0 {
		return nil, ErrInvalid
	}
	return s.repository.ListAdmin(ctx, status, limit, offset)
}

func (s *Service) Get(ctx context.Context, slug string) (Post, error) {
	slug = strings.TrimSpace(slug)
	if !validSlug(slug) {
		return Post{}, ErrNotFound
	}
	return s.repository.Get(ctx, slug)
}

func (s *Service) Create(ctx context.Context, authorID string, input Input) (Post, error) {
	input, err := normalize(input)
	if err != nil {
		return Post{}, err
	}
	if !validID(authorID) {
		return Post{}, ErrInvalid
	}
	return s.repository.Create(ctx, authorID, input)
}

func (s *Service) Update(ctx context.Context, postID string, input Input) (Post, error) {
	input, err := normalize(input)
	if err != nil {
		return Post{}, err
	}
	if !validID(postID) {
		return Post{}, ErrInvalid
	}
	return s.repository.Update(ctx, postID, input)
}

func (s *Service) SetStatus(ctx context.Context, postID, status string) (Post, error) {
	if !validID(postID) || (status != "published" && status != "draft" && status != "archived") {
		return Post{}, ErrInvalid
	}
	return s.repository.SetStatus(ctx, postID, status)
}

func (s *Service) Delete(ctx context.Context, postID string) error {
	if !validID(postID) {
		return ErrInvalid
	}
	return s.repository.Delete(ctx, postID)
}

func normalize(input Input) (Input, error) {
	input.Title = strings.TrimSpace(input.Title)
	input.Slug = strings.Trim(strings.ToLower(strings.TrimSpace(input.Slug)), "-")
	if input.Slug == "" {
		input.Slug = slugify(input.Title)
	}
	input.Excerpt = strings.TrimSpace(input.Excerpt)
	input.Body = strings.TrimSpace(input.Body)
	input.CoverURL = strings.TrimSpace(input.CoverURL)
	if input.Title == "" || len(input.Title) > 180 || !validSlug(input.Slug) ||
		len(input.Excerpt) > 500 || input.Body == "" || len(input.Body) > 200_000 ||
		len(input.CoverURL) > 2048 || len(input.Tags) > 20 {
		return Input{}, ErrInvalid
	}
	for index, tag := range input.Tags {
		input.Tags[index] = strings.TrimSpace(tag)
		if input.Tags[index] == "" || len(input.Tags[index]) > 40 {
			return Input{}, ErrInvalid
		}
	}
	return input, nil
}

func slugify(value string) string {
	var result strings.Builder
	dash := false
	for _, char := range strings.ToLower(value) {
		switch {
		case unicode.IsLetter(char) || unicode.IsDigit(char):
			result.WriteRune(char)
			dash = false
		case !dash && result.Len() > 0:
			result.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(result.String(), "-")
}

func validSlug(slug string) bool {
	if slug == "" || len(slug) > 180 || strings.HasPrefix(slug, "-") || strings.HasSuffix(slug, "-") {
		return false
	}
	for _, char := range slug {
		if !(char >= 'a' && char <= 'z' || char >= '0' && char <= '9' || char == '-') {
			return false
		}
	}
	return true
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
