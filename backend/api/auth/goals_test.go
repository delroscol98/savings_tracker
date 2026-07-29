package auth_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/api/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

var tokenSecret string

func TestCreateGoalHandler(t *testing.T) {
	userId := uuid.New()
	tokenSecret = "secret"
	expiresIn := time.Hour

	jwt, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

	tests := []struct {
		name         string
		body         io.Reader
		wantStatus   int
		wantErr      string
		setupHeaders func(*http.Request)
		setupMock    func(*mockDB)
	}{
		{
			name:       "valid goal",
			body:       strings.NewReader(fmt.Sprintf(`{"target": 1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusCreated,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			setupMock: func(md *mockDB) {},
		},
		{
			name:       "Authorization header not present",
			body:       strings.NewReader(fmt.Sprintf(`{"target": 1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "authorization header not present",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
			},
			setupMock: func(md *mockDB) {},
		},
		{
			name:       "malformed bearer token",
			body:       strings.NewReader(fmt.Sprintf(`{"target": 1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "malformed header",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", "Bear ")
			},
			setupMock: func(md *mockDB) {},
		},
		{
			name:       "Unable to parse jwt",
			body:       strings.NewReader(fmt.Sprintf(`{"target": 1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "error parsing jwt token",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", "Bearer ")
			},
			setupMock: func(md *mockDB) {},
		},
		{
			name:       "Poorly formatted body",
			body:       strings.NewReader(fmt.Sprintf(`{target": 1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Error decoding body",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			setupMock: func(md *mockDB) {},
		},
		{
			name:       "Invalid parameters",
			body:       strings.NewReader(fmt.Sprintf(`{"target": -1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new goal",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			setupMock: func(md *mockDB) {},
		},
		{
			name:       "database error",
			body:       strings.NewReader(fmt.Sprintf(`{"target": 1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Error creating goal",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			setupMock: func(md *mockDB) {
				md.CreateGoalErr = errors.New("Goal table error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDB{
				Goals: make(map[string]database.Goal),
			}
			tt.setupMock(md)

			api := auth.AuthConfig{Queries: md}
			api.JWTSecret = tokenSecret
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/goals", tt.body)
			tt.setupHeaders(r)
			api.CreateGoalHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf(`
Expected status code: %v
Actual status code:   %v
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
		})
	}
}

func TestDeleteGoalHandler(t *testing.T) {
	// Setup JWT
	userId := uuid.New()
	tokenSecret = "secret"
	expiresIn := time.Hour

	jwt, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

	// Setup goalId
	goalId := uuid.New()

	tests := []struct {
		name         string
		goalId       string
		wantStatus   int
		wantErr      string
		setupHeaders func(*http.Request)
		seedDB       func(*mockDB)
	}{
		{
			name:       "valid delete",
			goalId:     goalId.String(),
			wantStatus: http.StatusOK,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:        goalId,
					Target:    1000,
					Deadline:  time.Now().AddDate(1, 0, 0),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					UserID:    userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "Unable to parse goalID",
			goalId:     "Invalid",
			wantStatus: http.StatusBadRequest,
			wantErr:    "error parsing goal id",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:        goalId,
					Target:    1000,
					Deadline:  time.Now().AddDate(1, 0, 0),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					UserID:    userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "Goal not found",
			goalId:     goalId.String(),
			wantStatus: http.StatusNotFound,
			wantErr:    "error finding goal",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			seedDB: func(md *mockDB) {},
		},
		{
			name:       "mismatch user id",
			goalId:     goalId.String(),
			wantStatus: http.StatusForbidden,
			wantErr:    "mismatch user id",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:        goalId,
					Target:    1000,
					Deadline:  time.Now().AddDate(1, 0, 0),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					UserID:    uuid.New(),
				}

				md.Goals[goalId.String()] = goal
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDB{
				Goals: make(map[string]database.Goal),
			}
			tt.seedDB(md)
			api := auth.AuthConfig{Queries: md}
			api.JWTSecret = tokenSecret
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/goals/%v", tt.goalId), nil)
			r.SetPathValue("goalId", tt.goalId)
			tt.setupHeaders(r)
			api.DeleteGoalHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf(`
Expected status code: %v
Actual status code:   %v
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
		})
	}
}
