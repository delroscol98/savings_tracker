package handlers

import (
	"context"
	"sync/atomic"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

type Database interface {
	Ping(ctx context.Context) (int32, error)
	CreateUser(ctx context.Context, email string) (database.User, error)
}

type ApiConfig struct {
	FileserverHits  atomic.Int32
	DatabaseQueries Database
}
