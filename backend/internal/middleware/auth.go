package middleware

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

type userIdContextKey string

const userIdKey userIdContextKey = "userId"

func WithUserId(ctx context.Context, id uuid.UUID) context.Context {
	return context.WithValue(ctx, userIdKey, id)
}

func GetUserId(ctx context.Context) (uuid.UUID, bool) {
	userId, ok := ctx.Value(userIdKey).(uuid.UUID)
	return userId, ok
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// JWT
		secret := os.Getenv("JWT_SECRET")

		// JWT validation
		token, err := auth.GetBearerToken(r.Header)
		if err != nil {
			log.Print(err)
			response.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		userId, err := auth.ValidateJWT(token, secret)
		if err != nil {
			log.Print(err)
			response.RespondWithError(w, http.StatusUnauthorized, err.Error())
			return
		}

		ctx := WithUserId(r.Context(), userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
