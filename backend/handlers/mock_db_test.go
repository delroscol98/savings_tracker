package handlers_test

import (
	"context"
	"time"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type mockDB struct {
	pingErr       error
	CreateUserErr error
	users         map[string]database.User
}

func (m *mockDB) Ping(ctx context.Context) (int32, error) {
	return 1, m.pingErr
}

func (m *mockDB) CreateUser(ctx context.Context, email string) (database.User, error) {
	if m.CreateUserErr != nil {
		return database.User{}, m.CreateUserErr
	}

	if m.users == nil {
		m.users = make(map[string]database.User)
	}

	_, ok := m.users[email]
	if ok {
		return database.User{}, &pq.Error{Code: pqerror.Code("23505")}
	}

	user := database.User{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Email:     email,
	}
	m.users[email] = user

	return user, nil
}
