package worker

import (
	"context"

	"github.com/harshpn/taskflow/internal/store"
)

// StoreUserLookup adapts *store.Store to the UserLookup interface.
type StoreUserLookup struct {
	store *store.Store
}

func NewStoreUserLookup(s *store.Store) *StoreUserLookup {
	return &StoreUserLookup{store: s}
}

func (a *StoreUserLookup) GetUserByID(ctx context.Context, userID string) (name, email string, err error) {
	user, err := a.store.GetUserByID(ctx, userID)
	if err != nil {
		return "", "", err
	}
	return user.Name, user.Email, nil
}
