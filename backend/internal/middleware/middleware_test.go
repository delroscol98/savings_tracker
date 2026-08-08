package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
)

var apiHits atomic.Int32

func TestMiddlewareMetricInc(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := middleware.MetricInc(&apiHits, next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/app/", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusTeapot {
		t.Errorf(`
Expecting status code: %v
Actual status code:    %v
`, http.StatusTeapot, w.Code)
	}

	hits := apiHits.Load()
	if hits != 1 {
		t.Errorf(`
Expecting FileserverHits: 1
Actual FileserverHits:    %v
`, hits)
	}
}

func TestRequireAuth(t *testing.T) {
	t.Setenv("JWT_SECRET", "secret")

	userID := uuid.New()
	validToken, err := auth.MakeJWT(userID, "secret", time.Hour)
	if err != nil {
		t.Fatalf("failed to create valid JWT: %v", err)
	}
	expiredToken, err := auth.MakeJWT(userID, "secret", -time.Hour)
	if err != nil {
		t.Fatalf("failed to create expired JWT: %v", err)
	}

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantErr    string
		wantUserID bool
	}{
		{
			name:       "valid token",
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusOK,
			wantUserID: true,
		},
		{
			name:       "authorization header not present",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "authorization header not present",
		},
		{
			name:       "malformed header",
			authHeader: "Bear ",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "malformed header",
		},
		{
			name:       "invalid signature",
			authHeader: "Bearer invalid.token.string",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "error parsing jwt token",
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + expiredToken,
			wantStatus: http.StatusUnauthorized,
			wantErr:    "token is expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotUserID uuid.UUID
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotUserID, _ = middleware.GetUserId(r.Context())
				w.WriteHeader(http.StatusOK)
			})

			handler := middleware.RequireAuth(next)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/goals", nil)
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}
			handler.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf(`
Expecting status code: %v
Actual status code:    %v
`, tt.wantStatus, w.Code)
			}

			if tt.wantErr != "" {
				var body map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				if body["error"] != tt.wantErr {
					t.Errorf("want error %q, got %q", tt.wantErr, body["error"])
				}
			}

			if tt.wantUserID && gotUserID != userID {
				t.Errorf("want user id %v in context, got %v", userID, gotUserID)
			}
		})
	}
}
