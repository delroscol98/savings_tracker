package goals_test

import (
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
	"github.com/google/uuid"
)

func withUserContext(r *http.Request, userId uuid.UUID) *http.Request {
	ctx := middleware.WithUserId(r.Context(), userId)
	return r.WithContext(ctx)
}
