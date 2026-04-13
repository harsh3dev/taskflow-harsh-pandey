package service

import (
	"context"

	"github.com/harshpn/taskflow/internal/store"
)

type UserService struct {
	store *store.Store
}

func NewUserService(store *store.Store) *UserService {
	return &UserService{store: store}
}

func (s *UserService) ListUsers(ctx context.Context, search string) ([]User, error) {
	users, err := s.store.ListUsers(ctx, search)
	if err != nil {
		return nil, err
	}
	return usersFromStore(users), nil
}
