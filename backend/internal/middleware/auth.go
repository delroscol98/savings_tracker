package middleware

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

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

		ctx := context.WithValue(r.Context(), "userId", userId)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
