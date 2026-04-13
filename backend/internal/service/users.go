package service

import (
	"context"
)

type UserService struct {
	repo userListStore
}

func NewUserService(repo userListStore) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) ListUsers(ctx context.Context, search string) ([]User, error) {
	users, err := s.repo.ListUsers(ctx, search)
	if err != nil {
		return nil, err
	}
	return usersFromStore(users), nil
}
