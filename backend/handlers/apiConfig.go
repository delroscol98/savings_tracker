package handlers

import (
	"sync/atomic"

	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

type ApiConfig struct {
	FileserverHits  atomic.Int32
	DatabaseQueries *database.Queries
}
