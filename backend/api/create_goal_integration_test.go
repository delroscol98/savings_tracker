package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/api/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
)

func TestCreateGoalHandler_Integration(t *testing.T) {
	expiresIn := time.Hour

	// Seed Database with a single user
	hashedPw, err := auth.HashPassword("ThisIsATestPassword")
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
		Email:          "foobar@foobar.com",
		HashedPassword: hashedPw,
		FullName:       "Foo bar",
	})
	if err != nil {
		t.Fatalf("Failed to seed new user: %v", err)
	}

	tests := []struct {
		name         string
		target       int32
		deadline     string
		userId       uuid.UUID
		wantStatus   int
		wantErr      string
		setupHeaders func(*http.Request)
		checkGoal    func(*testing.T, *auth.Goal)
	}{
		{
			name:       "valid goal",
			target:     1000,
			deadline:   time.Now().Add(time.Hour).Format(time.RFC3339),
			userId:     user.ID,
			wantStatus: http.StatusCreated,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				jwt, _ := auth.MakeJWT(user.ID, JWTSecret, expiresIn)

				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			checkGoal: func(t *testing.T, g *auth.Goal) {
				if g.Id == uuid.Nil {
					t.Error("Goal ID should not be zero-value")
				}
				if g.CreatedAt.IsZero() {
					t.Error("Goal CreatedAt timestamp should not be zero-value")
				}
				if g.UpdatedAt.IsZero() {
					t.Error("Goal UpdatedAt timestamp should not be zero-value")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := strings.NewReader(fmt.Sprintf(`{"target": %v, "deadline": "%s", "user_id": "%v"}`, tt.target, tt.deadline, tt.userId))
			req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/goals", body)
			if err != nil {
				t.Fatalf("Error creating new request: %v", err)
			}
			tt.setupHeaders(req)
			resp, err := http.DefaultClient.Do(req)

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

			if tt.checkGoal != nil {
				goal := auth.Goal{}
				decoder := json.NewDecoder(resp.Body)
				err := decoder.Decode(&goal)
				if err != nil {
					t.Fatalf("Failed to decode goal: %v", err)
				}

				if goal.UserId != tt.userId {
					t.Errorf(`
Expected user ID: %v,
Actual user ID:   %v
`, tt.userId, goal.UserId)
				}

				tt.checkGoal(t, &goal)
			}
		})
	}
}
