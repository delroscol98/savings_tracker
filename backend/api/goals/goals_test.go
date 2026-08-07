package goals_test

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

	"github.com/delroscol98/savings_tracker/backend/api/goals"
	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

var tokenSecret string

func TestGetGoalsHandler(t *testing.T) {
	goalId1 := uuid.New()
	goalId2 := uuid.New()
	userId := uuid.New()
	otherUserId := uuid.New()
	target1 := int32(1000)
	target2 := int32(5000)
	deadline1 := time.Now().AddDate(0, 6, 0)
	deadline2 := time.Now().AddDate(1, 0, 0)

	tokenSecret = "secret"
	expiresIn := time.Hour
	token, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

	tests := []struct {
		name         string
		wantStatus   int
		wantErr      string
		setupHeaders func(*http.Request)
		seedDB       func(*mockDB)
		checkGoals   func(*testing.T, []goals.Goal)
	}{
		{
			name:       "valid goals",
			wantStatus: http.StatusOK,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				goal1 := database.Goal{
					ID:       goalId1,
					Target:   target1,
					Deadline: deadline1,
					UserID:   userId,
				}
				goal2 := database.Goal{
					ID:       goalId2,
					Target:   target2,
					Deadline: deadline2,
					UserID:   userId,
				}

				md.Goals[goalId1.String()] = goal1
				md.Goals[goalId2.String()] = goal2
			},
			checkGoals: func(t *testing.T, goals []goals.Goal) {
				if len(goals) != 2 {
					t.Errorf("want %v goals, got %v", 2, len(goals))
				}

				for _, goal := range goals {
					switch goal.Id {
					case goalId1:
						if goal.Target != target1 {
							t.Errorf("want goal 1 target %v, got %v", target1, goal.Target)
						}
						if goal.Deadline.Compare(deadline1) != 0 {
							t.Errorf("goal 1 deadline should match seed")
						}
						if goal.UserId != userId {
							t.Errorf("goal 1 user id mismatch")
						}
					case goalId2:
						if goal.Target != target2 {
							t.Errorf("want goal 2 target %v, got %v", target2, goal.Target)
						}
						if goal.Deadline.Compare(deadline2) != 0 {
							t.Errorf("goal 2 deadline should match seed")
						}
						if goal.UserId != userId {
							t.Errorf("goal 2 user id mismatch")
						}
					default:
						t.Errorf("unexpected goal id %v", goal.Id)
					}
				}
			},
		},
		{
			name:       "no goals",
			wantStatus: http.StatusOK,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {},
			checkGoals: func(t *testing.T, goals []goals.Goal) {
				if len(goals) != 0 {
					t.Errorf("want %v goals, got %v", 0, len(goals))
				}
			},
		},
		{
			name:       "mismatched user id",
			wantStatus: http.StatusOK,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				otherToken, _ := auth.MakeJWT(otherUserId, tokenSecret, expiresIn)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", otherToken))
			},
			seedDB: func(md *mockDB) {
				goal1 := database.Goal{
					ID:       goalId1,
					Target:   target1,
					Deadline: deadline1,
					UserID:   userId,
				}

				md.Goals[goalId1.String()] = goal1
			},
			checkGoals: func(t *testing.T, goals []goals.Goal) {
				if len(goals) != 0 {
					t.Errorf("want %v goals, got %v", 0, len(goals))
				}
			},
		},
		{
			name:       "authorization header not present",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "authorization header not present",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
			},
			seedDB: func(md *mockDB) {},
		},
		{
			name:       "malformed header",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "malformed header",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", "Bear ")
			},
			seedDB: func(md *mockDB) {},
		},
		{
			name:       "unable to parse token",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "error parsing jwt token",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", "Bearer invalid.token.string")
			},
			seedDB: func(md *mockDB) {},
		},
		{
			name:       "token is expired",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "token is expired",
			setupHeaders: func(r *http.Request) {
				expiredToken, _ := auth.MakeJWT(userId, tokenSecret, -expiresIn)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", expiredToken))
			},
			seedDB: func(md *mockDB) {},
		},
		{
			name:       "database error",
			wantStatus: http.StatusInternalServerError,
			wantErr:    "error fetching goals",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				md.GetGoalsErr = errors.New("Goal table error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDB{
				Goals: make(map[string]database.Goal),
			}
			tt.seedDB(md)

			api := goals.GoalsConfig{Queries: md}
			api.JWTSecret = tokenSecret
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodGet, "/api/goals", nil)
			tt.setupHeaders(r)
			api.GetGoalsHandler(w, r)

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

			if tt.checkGoals != nil {
				if tt.name == "no goals" {
					raw := w.Body.String()
					if !strings.HasPrefix(raw, "[") || !strings.HasSuffix(strings.TrimSpace(raw), "]") {
						t.Errorf("expected raw body to be an array, got %q", raw)
					}
				}

				var goals []goals.Goal
				if err := json.NewDecoder(w.Body).Decode(&goals); err != nil {
					t.Fatalf("failed to decode goals: %v", err)
				}
				tt.checkGoals(t, goals)
			}
		})
	}
}

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

			api := goals.GoalsConfig{Queries: md}
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
		checkGoal    func(*testing.T, goals.Goal)
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
			checkGoal: func(t *testing.T, g goals.Goal) {
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
			api := goals.GoalsConfig{Queries: &md}
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
				var goal goals.Goal
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
			api := goals.GoalsConfig{Queries: md}
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
