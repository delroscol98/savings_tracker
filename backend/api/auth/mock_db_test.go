package auth_test

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/delroscol98/savings_tracker/backend/api/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/lib/pq/pqerror"
)

type mockDB struct {
	CreateUserErr                  error
	LoginErr                       error
	GetUserByEmailErr              error
	CreatePasswordResetTokenErr    error
	GetPasswordResetTokenByHashErr error
	ConsumePasswordResetTokenErr   error
	DeactivateUserTokensErr        error
	UpdateUserPasswordErr          error
	users                          map[string]database.User
	PasswordResetTokens            map[string]database.PasswordResetToken
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
		FullName:       params.FullName,
	}
	m.users[params.Email] = user

	return user, nil
}

func (m *mockDB) Login(context context.Context, email string) (database.User, error) {
	if m.LoginErr != nil {
		return database.User{}, m.LoginErr
	}

	password1 := "ThisIsATestPassword"
	password1Hash, _ := auth.HashPassword(password1)

	if m.users == nil {
		m.users = map[string]database.User{
			"test-1@example.com": {
				ID:             uuid.New(),
				CreatedAt:      time.Now(),
				UpdatedAt:      time.Now(),
				Email:          "test-1@example.com",
				HashedPassword: password1Hash,
			},
		}
	}

	user, ok := m.users[email]
	if !ok {
		return database.User{}, sql.ErrNoRows
	}

	return user, nil
}

func (m *mockDB) GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error) {
	if m.GetUserByEmailErr != nil {
		return database.GetUserByEmailRow{}, m.GetUserByEmailErr
	}

	user, ok := m.users[email]
	if !ok {
		return database.GetUserByEmailRow{}, errors.New("User not found")
	}

	return database.GetUserByEmailRow{
		ID:    user.ID,
		Email: user.Email,
	}, nil
}

func (m *mockDB) CreatePasswordResetToken(ctx context.Context, params database.CreatePasswordResetTokenParams) (database.PasswordResetToken, error) {
	if m.CreatePasswordResetTokenErr != nil {
		return database.PasswordResetToken{}, m.CreatePasswordResetTokenErr
	}

	if m.PasswordResetTokens == nil {
		m.PasswordResetTokens = make(map[string]database.PasswordResetToken)
	}

	passwordResetToken := database.PasswordResetToken{
		ID:        uuid.New(),
		UserID:    params.UserID,
		TokenHash: params.TokenHash,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(15 * time.Minute),
	}
	m.PasswordResetTokens[params.TokenHash] = passwordResetToken

	return passwordResetToken, nil
}

func (m *mockDB) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (database.PasswordResetToken, error) {
	if m.GetPasswordResetTokenByHashErr != nil {
		return database.PasswordResetToken{}, m.GetPasswordResetTokenByHashErr
	}

	passwordResetToken, ok := m.PasswordResetTokens[tokenHash]
	if !ok {
		return database.PasswordResetToken{}, errors.New("Password reset token not found")
	}

	return passwordResetToken, nil
}

func (m *mockDB) ConsumePasswordResetToken(ctx context.Context, id uuid.UUID) error {
	if m.ConsumePasswordResetTokenErr != nil {
		return m.ConsumePasswordResetTokenErr
	}

	for _, passwordResetToken := range m.PasswordResetTokens {
		if id == passwordResetToken.ID {
			passwordResetToken.ConsumedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}
			m.PasswordResetTokens[passwordResetToken.TokenHash] = passwordResetToken
			return nil
		}
	}

	return errors.New("Password reset token not found")
}

func (m *mockDB) DeactivateUserTokens(ctx context.Context, userID uuid.UUID) error {
	if m.DeactivateUserTokensErr != nil {
		return m.DeactivateUserTokensErr
	}

	for _, passwordResetToken := range m.PasswordResetTokens {
		if userID == passwordResetToken.UserID && !passwordResetToken.ConsumedAt.Valid {
			passwordResetToken.ConsumedAt = sql.NullTime{
				Time:  time.Now(),
				Valid: true,
			}

			m.PasswordResetTokens[passwordResetToken.TokenHash] = passwordResetToken
			return nil
		}
	}

	return errors.New("Password reset token not found")
}

func (m *mockDB) UpdateUserPassword(ctx context.Context, params database.UpdateUserPasswordParams) error {
	if m.UpdateUserPasswordErr != nil {
		return m.UpdateUserPasswordErr
	}

	for _, user := range m.users {
		if params.ID == user.ID {
			user.HashedPassword = params.HashedPassword
			m.users[user.Email] = user

			return nil
		}
	}

	return errors.New("User not found")
}
