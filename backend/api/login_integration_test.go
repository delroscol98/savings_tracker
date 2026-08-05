package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/delroscol98/savings_tracker/backend/api/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

func TestLoginHandler_Integration(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		wantStatus int
		wantErr    string
		seedDB     func(*testing.T)
	}{
		{
			name:       "valid login",
			email:      "foo-1@example.com",
			password:   "AnotherTestPassword",
			wantStatus: http.StatusOK,
			wantErr:    "",
			seedDB: func(t *testing.T) {
				hashedPw, err := auth.HashPassword("AnotherTestPassword")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}

				_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-1@example.com",
					HashedPassword: hashedPw,
				})
				if err != nil {
					t.Fatalf("Failed to seed new user: %v", err)
				}
			},
		},
		{
			name:       "incorrect email",
			email:      "foo-3@example.com",
			password:   "AnotherTestPassword",
			wantStatus: http.StatusForbidden,
			wantErr:    "Incorrect email or password",
			seedDB: func(t *testing.T) {
				hashedPw, err := auth.HashPassword("ThisIsATestPassword")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}

				_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-2@example.com",
					HashedPassword: hashedPw,
				})
				if err != nil {
					t.Fatalf("Failed to seed new user: %v", err)
				}
			},
		},
		{
			name:       "incorrect password",
			email:      "foo-4@example.com",
			password:   "AnotherTestPassword",
			wantStatus: http.StatusForbidden,
			wantErr:    "Incorrect email or password",
			seedDB: func(t *testing.T) {
				hashedPw, err := auth.HashPassword("ThisIsATestPassword")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}

				_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-4@example.com",
					HashedPassword: hashedPw,
				})
				if err != nil {
					t.Fatalf("Failed to seed new user: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedDB(t)

			params := strings.NewReader(fmt.Sprintf(`{"email": "%v", "password": "%v"}`, tt.email, tt.password))
			resp, err := http.Post(testServer.URL+"/api/login", "application/json", params)
			if err != nil {
				t.Fatalf("Error sending post request: %v", err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Errorf(`
Expected status code: %v,
Actual status code:   %v
`, tt.wantStatus, resp.StatusCode)
			}

			if tt.wantErr != "" {
				body := make(map[string]interface{})
				decoder := json.NewDecoder(resp.Body)
				err := decoder.Decode(&body)
				if err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}

				if body["error"] != tt.wantErr {
					t.Errorf(`
Expected error: %v,
Actual error:   %v
`, tt.wantErr, body["error"])
				}
			}
		})
	}
}

func TestLoginExpiresIn_Integration(t *testing.T) {
	hashedPw, err := auth.HashPassword("AnotherTestPassword")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          "expiry-test@example.com",
		HashedPassword: hashedPw,
	})
	if err != nil {
		t.Fatalf("Failed to seed new user: %v", err)
	}

	tests := []struct {
		name      string
		expiresIn int64
		minTTL    time.Duration
		maxTTL    time.Duration
	}{
		{
			name:      "explicit expiry",
			expiresIn: 60,
			minTTL:    59 * time.Second,
			maxTTL:    61 * time.Second,
		},
		{
			name:      "default expiry",
			expiresIn: 0,
			minTTL:    3599 * time.Second,
			maxTTL:    3601 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := strings.NewReader(fmt.Sprintf(`{"email": "expiry-test@example.com", "password": "AnotherTestPassword", "expires_in": %v}`, tt.expiresIn))
			resp, err := http.Post(testServer.URL+"/api/login", "application/json", params)
			if err != nil {
				t.Fatalf("Error sending post request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Fatalf(`
Expected status code: %v
Actual status code:   %v
`, http.StatusOK, resp.StatusCode)
			}

			var user auth.User
			decoder := json.NewDecoder(resp.Body)
			err = decoder.Decode(&user)
			if err != nil {
				t.Fatalf("failed to decode response body: %v", err)
			}

			claims := &jwt.RegisteredClaims{}
			token, err := jwt.ParseWithClaims(user.Token, claims, func(token *jwt.Token) (any, error) {
				return []byte(JWTSecret), nil
			})
			if err != nil {
				t.Fatalf("failed to parse jwt token: %v", err)
			}
			if !token.Valid {
				t.Fatalf("expected token to be valid")
			}

			ttl := claims.ExpiresAt.Time.Sub(time.Now())
			if ttl < tt.minTTL || ttl > tt.maxTTL {
				t.Fatalf(`
Expected token TTL within window: [%v, %v]
Actual token TTL:                 %v
`, tt.minTTL, tt.maxTTL, ttl)
			}
		})
	}
}
