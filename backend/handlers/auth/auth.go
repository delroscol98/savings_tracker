package auth

import (
	"database/sql"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
)

type AuthConfig struct {
	Queries     handlers.Queries
	Database    *sql.DB
	RateLimiter *ratelimit.RateLimiter
	EmailSender handlers.EmailSender
}
