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
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestCreateGoalHandler_Integration(t *testing.T) {
	expiresIn := time.Hour
	extraUserId := uuid.New()

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
			deadline:   time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
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
			},
		},
		{
			name:       "User ID doesn't exist",
			target:     1000,
			deadline:   time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
			userId:     extraUserId,
			wantStatus: http.StatusBadRequest,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				jwt, _ := auth.MakeJWT(extraUserId, JWTSecret, expiresIn)

				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "Expired JWT",
			target:     1000,
			deadline:   time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
			userId:     user.ID,
			wantStatus: http.StatusUnauthorized,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Subject:   user.ID.String(),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
				})
				expiredJWT, _ := token.SignedString([]byte(JWTSecret))

				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", expiredJWT))
			},
		},
		{
			name:       "Mismatched user ID",
			target:     1000,
			deadline:   time.Now().AddDate(1, 0, 0).Format(time.RFC3339),
			userId:     user.ID,
			wantStatus: http.StatusUnauthorized,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
					Subject:   extraUserId.String(),
					ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					IssuedAt:  jwt.NewNumericDate(time.Now()),
				})
				expiredJWT, _ := token.SignedString([]byte(JWTSecret))

				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", expiredJWT))
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
			if err != nil {
				t.Fatalf("Error sending request: %v", err)
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

func TestDeleteGoalHandler_Integration(t *testing.T) {
	tests := []struct {
		name         string
		wantStatus   int
		wantErr      string
		seedDB       func(*testing.T) (database.User, database.Goal)
		setupRequest func(database.Goal, *testing.T) *http.Request
		setupHeaders func(database.User, *http.Request)
	}{
		{
			name:       "Valid delete",
			wantStatus: http.StatusOK,
			wantErr:    "",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				hash, _ := auth.HashPassword("ThisIsATestPassword")

				user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-bar2@example.com",
					HashedPassword: hash,
					FullName:       "John Smith",
				})
				if err != nil {
					t.Fatalf("error seeding user into database")
				}

				goal, err := dbQueries.CreateGoal(context.Background(), database.CreateGoalParams{
					Target:   1000,
					Deadline: time.Now().AddDate(1, 0, 0),
					UserID:   user.ID,
				})
				if err != nil {
					t.Fatalf("error seeding goal into database")
				}

				return user, goal
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/api/goals/%v", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, "secret", time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "goal not found",
			wantStatus: http.StatusNotFound,
			wantErr:    "error finding goal",
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/api/goals/%v", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				hash, _ := auth.HashPassword("ThisIsATestPassword")

				user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-bar3@example.com",
					HashedPassword: hash,
					FullName:       "John Smith",
				})
				if err != nil {
					t.Fatalf("error seeding user into database")
				}

				return user, database.Goal{}
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, "secret", time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "authorization header not found",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "authorization header not present",
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/api/goals/%v", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				hash, _ := auth.HashPassword("ThisIsATestPassword")

				user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-bar4@example.com",
					HashedPassword: hash,
					FullName:       "John Smith",
				})
				if err != nil {
					t.Log(err)
					t.Fatalf("error seeding user into database")
				}

				goal, err := dbQueries.CreateGoal(context.Background(), database.CreateGoalParams{
					Target:   1000,
					Deadline: time.Now().AddDate(1, 0, 0),
					UserID:   user.ID,
				})
				if err != nil {
					t.Fatalf("error seeding goal into database")
				}

				return user, goal
			},
			setupHeaders: func(u database.User, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
			},
		},
		{
			name:       "invalid goal id",
			wantStatus: http.StatusBadRequest,
			wantErr:    "error parsing goal id",
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/api/goals/abc", testServer.URL), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				hash, _ := auth.HashPassword("ThisIsATestPassword")

				user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-bar5@example.com",
					HashedPassword: hash,
					FullName:       "John Smith",
				})
				if err != nil {
					t.Fatalf("error seeding user into database")
				}

				goal, err := dbQueries.CreateGoal(context.Background(), database.CreateGoalParams{
					Target:   1000,
					Deadline: time.Now().AddDate(1, 0, 0),
					UserID:   user.ID,
				})
				if err != nil {
					t.Fatalf("error seeding goal into database")
				}

				return user, goal
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, "secret", time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "mismatch user id",
			wantStatus: http.StatusForbidden,
			wantErr:    "mismatch user id",
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/api/goals/%v", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				hash, _ := auth.HashPassword("ThisIsATestPassword")

				user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-bar7@example.com",
					HashedPassword: hash,
					FullName:       "John Smith",
				})
				if err != nil {
					t.Fatalf("error seeding user into database")
				}

				goal, err := dbQueries.CreateGoal(context.Background(), database.CreateGoalParams{
					Target:   1000,
					Deadline: time.Now().AddDate(1, 0, 0),
					UserID:   user.ID,
				})
				if err != nil {
					t.Fatalf("error seeding goal into database")
				}

				return user, goal
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(uuid.New(), "secret", time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "expired jwt",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "token is expired",
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%v/api/goals/%v", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				hash, _ := auth.HashPassword("ThisIsATestPassword")

				user, err := dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-bar8@example.com",
					HashedPassword: hash,
					FullName:       "John Smith",
				})
				if err != nil {
					t.Fatalf("error seeding user into database")
				}

				goal, err := dbQueries.CreateGoal(context.Background(), database.CreateGoalParams{
					Target:   1000,
					Deadline: time.Now().AddDate(1, 0, 0),
					UserID:   user.ID,
				})
				if err != nil {
					t.Fatalf("error seeding goal into database")
				}

				return user, goal
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(uuid.New(), "secret", -time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, goal := tt.seedDB(t)

			req := tt.setupRequest(goal, t)
			tt.setupHeaders(user, req)
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("error sending request: %v", err)
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
