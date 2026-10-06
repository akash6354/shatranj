package lessons

import (
	"context"
	"encoding/json"
	"fmt"
)

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, userID string) ([]Course, error) {
	if userID == "" {
		return nil, ErrInvalidInput
	}
	return s.repository.List(ctx, userID)
}

func (s *Service) Get(ctx context.Context, userID, courseID string) (Course, error) {
	if userID == "" || !validLessonID(courseID) {
		return Course{}, ErrNotFound
	}
	return s.repository.Get(ctx, userID, courseID)
}

func (s *Service) UpdateProgress(ctx context.Context, userID, chapterID string, input ProgressInput) (Chapter, error) {
	if userID == "" || !validLessonID(chapterID) || input.ProgressPercent < 0 || input.ProgressPercent > 100 {
		return Chapter{}, fmt.Errorf("%w: valid chapter and progress percentage are required", ErrInvalidInput)
	}
	if len(input.LastPosition) > 0 && !json.Valid(input.LastPosition) {
		return Chapter{}, fmt.Errorf("%w: last_position must be valid JSON", ErrInvalidInput)
	}
	if len(input.LastPosition) > 0 {
		var lastPosition map[string]json.RawMessage
		if err := json.Unmarshal(input.LastPosition, &lastPosition); err != nil || lastPosition == nil {
			return Chapter{}, fmt.Errorf("%w: last_position must be a JSON object", ErrInvalidInput)
		}
	}
	if input.Completed {
		input.ProgressPercent = 100
	}
	input.Completed = input.Completed || input.ProgressPercent == 100
	return s.repository.UpdateProgress(ctx, userID, chapterID, input)
}

func validLessonID(id string) bool {
	if len(id) != 36 {
		return false
	}
	for index, value := range id {
		switch index {
		case 8, 13, 18, 23:
			if value != '-' {
				return false
			}
		default:
			if !(value >= '0' && value <= '9' || value >= 'a' && value <= 'f' || value >= 'A' && value <= 'F') {
				return false
			}
		}
	}
	return true
}
