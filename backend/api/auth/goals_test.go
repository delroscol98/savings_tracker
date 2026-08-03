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

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/api/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

var tokenSecret string

func TestCreateGoalHandler(t *testing.T) {
	userId := uuid.New()
	tokenSecret = "secret"
	expiresIn := time.Hour

	token, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

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
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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
			name:       "Unable to parse token",
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
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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

func TestUpdateGoalHandler(t *testing.T) {
	// Setup token
	userId := uuid.New()
	tokenSecret = "secret"
	expiresIn := time.Hour

	token, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

	// Static goal details
	goalId := uuid.New()

	// old goal
	target := int32(1000)
	deadline := time.Now().AddDate(0, 6, 0)

	// Setup updated goal
	updatedTarget := int32(5000)
	updatedDeadline := time.Now().AddDate(1, 0, 0)

	tests := []struct {
		name         string
		body         string
		goalId       string
		target       int32
		deadline     string
		wantStatus   int
		wantErr      string
		setupHeaders func(*testing.T, *http.Request)
		seedDB       func(*mockDB)
		checkGoal    func(*testing.T, auth.Goal)
	}{
		{
			name:       "valid update",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusOK,
			wantErr:    "",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
			checkGoal: func(t *testing.T, g auth.Goal) {
				// check updated target
				if g.Target != updatedTarget {
					t.Errorf(`
Expected updated target: %v
Actual updated target:   %v
`, updatedTarget, g.Target)
				}

				// check updated deadline
				if g.Deadline.Compare(deadline) != 1 {
					t.Errorf("goal deadline should be updated")
				}
			},
		},
		{
			name:       "unable to get bearer token",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "authorization header not present",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "malformed error",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "malformed header",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", "Invalid token")
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "token in unverifiable",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "token is unverifiable",
			setupHeaders: func(t *testing.T, r *http.Request) {
				claims := jwt.RegisteredClaims{
					Issuer:    "savings-tracker-access",
					IssuedAt:  jwt.NewNumericDate(time.Now().UTC()),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiresIn)),
					Subject:   userId.String(),
				}
				token := jwt.NewWithClaims(jwt.SigningMethodHS384, claims)
				tokenStr, err := token.SignedString([]byte(tokenSecret))
				if err != nil {
					t.Fatalf("Unable to sign JWT")
				}

				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", tokenStr))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "expired token",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "token is expired",
			setupHeaders: func(t *testing.T, r *http.Request) {
				token, _ := auth.MakeJWT(userId, tokenSecret, -time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "unable to parse token",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusUnauthorized,
			wantErr:    "error parsing jwt token",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", "Bearer invalid.token.string")
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "invalid goal id",
			goalId:     "Invalid",
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
			wantErr:    "error parsing goal id",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "goal not found",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusNotFound,
			wantErr:    "error finding goal",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {},
		},
		{
			name:       "mismatch user id",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusForbidden,
			wantErr:    "mismatch user id",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   uuid.New(),
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "poorly formatted body",
			body:       fmt.Sprintf(`{target": %v, "deadline": "%s"}`, updatedTarget, updatedDeadline.Format(time.RFC3339)),
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
			wantErr:    "error decoding body",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "invalid parameters",
			goalId:     goalId.String(),
			target:     -100,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new goal",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
			},
		},
		{
			name:       "database error",
			goalId:     goalId.String(),
			target:     updatedTarget,
			deadline:   updatedDeadline.Format(time.RFC3339),
			wantStatus: http.StatusBadRequest,
			wantErr:    "error updating goal",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal := database.Goal{
					ID:       goalId,
					Target:   target,
					Deadline: deadline,
					UserID:   userId,
				}

				md.Goals[goalId.String()] = goal
				md.UpdateGoalErr = errors.New("Goal table error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := mockDB{
				Goals: make(map[string]database.Goal),
			}
			tt.seedDB(&md)
			api := auth.AuthConfig{Queries: &md}
			api.JWTSecret = tokenSecret
			body := tt.body
			if body == "" {
				body = fmt.Sprintf(`{"target": %v, "deadline": "%s"}`, tt.target, tt.deadline)
			}

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodPut,
				fmt.Sprintf("/api/goals/%v", tt.goalId),
				strings.NewReader(body),
			)
			r.SetPathValue("goalId", tt.goalId)
			tt.setupHeaders(t, r)
			api.UpdateGoalHandler(w, r)

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

			if tt.checkGoal != nil {
				var goal auth.Goal
				if err := json.NewDecoder(w.Body).Decode(&goal); err != nil {
					t.Fatalf("failed to decode goal: %v", err)
				}
				tt.checkGoal(t, goal)
			}
		})
	}
}

func TestDeleteGoalHandler(t *testing.T) {
	// Setup token
	userId := uuid.New()
	tokenSecret = "secret"
	expiresIn := time.Hour

	token, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

	// Setup goalId
	goalId := uuid.New()

	tests := []struct {
		name         string
		goalId       string
		wantStatus   int
		wantErr      string
		setupHeaders func(*testing.T, *http.Request)
		seedDB       func(*mockDB)
	}{
		{
			name:       "valid delete",
			goalId:     goalId.String(),
			wantStatus: http.StatusOK,
			wantErr:    "",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {},
		},
		{
			name:       "mismatch user id",
			goalId:     goalId.String(),
			wantStatus: http.StatusForbidden,
			wantErr:    "mismatch user id",
			setupHeaders: func(t *testing.T, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
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
			tt.setupHeaders(t, r)
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
