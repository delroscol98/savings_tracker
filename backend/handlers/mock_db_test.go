package handlers_test

import (
	"context"
	"time"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
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

func (m *mockDB) CreateUser(ctx context.Context, params database.CreateUserParams) (database.User, error) {
	if m.CreateUserErr != nil {
		return database.User{}, m.CreateUserErr
	}

	if m.users == nil {
		m.users = make(map[string]database.User)
	}

	_, ok := m.users[params.Email]
	if ok {
		return database.User{}, &pq.Error{Code: pqerror.Code("23505")}
	}

	hashedPassword, _ := auth.HashPassword(params.HashedPassword)

	user := database.User{
		ID:             uuid.New(),
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Email:          params.Email,
		HashedPassword: hashedPassword,
	}
	m.users[params.Email] = user

	return user, nil
}
