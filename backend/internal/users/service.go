package users

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("user not found")

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) Get(ctx context.Context, id string) (User, error) {
	return s.repository.FindByID(ctx, id)
}
