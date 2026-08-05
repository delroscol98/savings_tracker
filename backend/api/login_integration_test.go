package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

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
