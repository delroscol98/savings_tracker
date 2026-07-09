package handlers

import (
	"context"
	"database/sql"
	"sync/atomic"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
	"github.com/google/uuid"
)

type Database interface {
	Ping(ctx context.Context) (int32, error)
	CreateUser(ctx context.Context, params database.CreateUserParams) (database.User, error)
	GetUserByEmail(ctx context.Context, email string) (database.GetUserByEmailRow, error)
	Login(ctx context.Context, email string) (database.User, error)
	CreatePasswordResetToken(ctx context.Context, params database.CreatePasswordResetTokenParams) (database.PasswordResetToken, error)
	GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (database.PasswordResetToken, error)
	ConsumePasswordResetToken(ctx context.Context, id uuid.UUID) error
	DeactiveUserTokens(ctx context.Context, userID uuid.UUID) error
	UpdateUserPassword(ctx context.Context, params database.UpdateUserPasswordParams) error
}

type ApiConfig struct {
	FileserverHits  atomic.Int32
	DatabaseQueries Database
	db              *sql.DB
	RateLimiter     *ratelimit.RateLimiter
}
