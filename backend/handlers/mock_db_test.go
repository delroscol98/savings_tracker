package handlers_test

import (
	"context"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

type mockDB struct {
	pingErr error
}

func (m *mockDB) Ping(ctx context.Context) (int32, error) {
	return 1, m.pingErr
}

func (m *mockDB) CreateUser(ctx context.Context, email string) (database.User, error) {
	return database.User{}, nil
}
