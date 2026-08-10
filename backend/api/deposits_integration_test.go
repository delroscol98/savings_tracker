package api_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/api/goals"
	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
)

func TestCreateDepositHandler_Integration(t *testing.T) {
	amount := int32(100)
	note := "test deposit"

	tests := []struct {
		name         string
		wantStatus   int
		wantErr      string
		seedDB       func(*testing.T) (database.User, database.Goal)
		setupRequest func(database.Goal, *testing.T) *http.Request
		setupHeaders func(database.User, *http.Request)
		checkDeposit func(*testing.T, goals.Deposit)
	}{
		{
			name:       "valid deposit",
			wantStatus: http.StatusCreated,
			wantErr:    "",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUserWithGoal(t, "foo-bar-deposit1@example.com", 1000, time.Now().AddDate(1, 0, 0))
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), strings.NewReader(fmt.Sprintf(`{"amount": %v, "note": "%v"}`, amount, note)))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			checkDeposit: func(t *testing.T, d goals.Deposit) {
				if d.Id == uuid.Nil {
					t.Error("Deposit ID should not be zero-value")
				}

				if d.Amount != amount {
					t.Errorf(`
Expected amount: %v
Actual amount:   %v
`, amount, d.Amount)
				}

				if d.Note != note {
					t.Errorf(`
Expected note: %v
Actual note:   %v
`, note, d.Note)
				}
			},
		},
		{
			name:       "goal not found",
			wantStatus: http.StatusNotFound,
			wantErr:    "error finding goal",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUser(t, "foo-bar-deposit2@example.com"), database.Goal{}
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), strings.NewReader(fmt.Sprintf(`{"amount": %v, "note": "%v"}`, amount, note)))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "mismatch user id",
			wantStatus: http.StatusForbidden,
			wantErr:    "mismatch user id",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUserWithGoal(t, "foo-bar-deposit3@example.com", 1000, time.Now().AddDate(1, 0, 0))
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), strings.NewReader(fmt.Sprintf(`{"amount": %v, "note": "%v"}`, amount, note)))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(uuid.New(), JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "invalid goal id",
			wantStatus: http.StatusBadRequest,
			wantErr:    "error parsing goal id",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUserWithGoal(t, "foo-bar-deposit4@example.com", 1000, time.Now().AddDate(1, 0, 0))
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/abc/deposits", testServer.URL), strings.NewReader(fmt.Sprintf(`{"amount": %v, "note": "%v"}`, amount, note)))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "authorization header not present",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "authorization header not present",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUserWithGoal(t, "foo-bar-deposit5@example.com", 1000, time.Now().AddDate(1, 0, 0))
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), strings.NewReader(fmt.Sprintf(`{"amount": %v, "note": "%v"}`, amount, note)))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
			},
		},
		{
			name:       "expired jwt",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "token is expired",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUserWithGoal(t, "foo-bar-deposit6@example.com", 1000, time.Now().AddDate(1, 0, 0))
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), strings.NewReader(fmt.Sprintf(`{"amount": %v, "note": "%v"}`, amount, note)))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(uuid.New(), JWTSecret, -time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "invalid parameters",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new deposit",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUserWithGoal(t, "foo-bar-deposit7@example.com", 1000, time.Now().AddDate(1, 0, 0))
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), strings.NewReader(`{"amount": -100, "note": "test deposit"}`))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "poorly formatted body",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Error decoding body",
			seedDB: func(t *testing.T) (database.User, database.Goal) {
				return seedUserWithGoal(t, "foo-bar-deposit8@example.com", 1000, time.Now().AddDate(1, 0, 0))
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), strings.NewReader(`{"amount": 100`))
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
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
				return
			}

			if tt.checkDeposit != nil {
				deposit := goals.Deposit{}
				decoder := json.NewDecoder(resp.Body)
				err := decoder.Decode(&deposit)
				if err != nil {
					t.Fatalf("failed to decode deposit: %v", err)
				}

				tt.checkDeposit(t, deposit)
			}
		})
	}
}

func TestGetDepositsByGoalAndUserHandler_Integration(t *testing.T) {
	tests := []struct {
		name          string
		wantStatus    int
		wantErr       string
		seedDB        func(*testing.T) (database.User, database.Goal, []database.Deposit)
		setupRequest  func(database.Goal, *testing.T) *http.Request
		setupHeaders  func(database.User, *http.Request)
		checkDeposits func(*testing.T, []goals.Deposit, []database.Deposit)
	}{
		{
			name:       "valid deposits",
			wantStatus: http.StatusOK,
			wantErr:    "",
			seedDB: func(t *testing.T) (database.User, database.Goal, []database.Deposit) {
				user, goal := seedUserWithGoal(t, "foo-bar-getdeposit1@example.com", 1000, time.Now().AddDate(1, 0, 0))

				deposit1 := seedDeposit(t, goal.ID, user.ID, 100, "test deposit")
				deposit2 := seedDeposit(t, goal.ID, user.ID, 200, "AnotherTest")

				otherGoal := seedGoal(t, user.ID, 500, time.Now().AddDate(1, 0, 0))
				seedDeposit(t, otherGoal.ID, user.ID, 999, "other goal deposit")

				return user, goal, []database.Deposit{deposit1, deposit2}
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			checkDeposits: func(t *testing.T, deposits []goals.Deposit, seeded []database.Deposit) {
				if len(deposits) != len(seeded) {
					t.Fatalf("Expected %v deposits, got %v", len(seeded), len(deposits))
				}

				for _, want := range seeded {
					found := false
					for _, got := range deposits {
						if got.Id == want.ID {
							found = true
							if got.Amount != want.Amount {
								t.Errorf(`
Expected amount: %v
Actual amount:   %v
`, want.Amount, got.Amount)
							}
							if got.Note != want.Note.String {
								t.Errorf(`
Expected note: %v
Actual note:   %v
`, want.Note.String, got.Note)
							}
						}
					}
					if !found {
						t.Errorf("seeded deposit %v not found in response", want.ID)
					}
				}
			},
		},
		{
			name:       "goal with no deposits",
			wantStatus: http.StatusOK,
			wantErr:    "",
			seedDB: func(t *testing.T) (database.User, database.Goal, []database.Deposit) {
				user, goal := seedUserWithGoal(t, "foo-bar-getdeposit2@example.com", 1000, time.Now().AddDate(1, 0, 0))

				return user, goal, nil
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
			checkDeposits: func(t *testing.T, deposits []goals.Deposit, seeded []database.Deposit) {
				if deposits == nil {
					t.Error("expected empty array, got null")
				}

				if len(deposits) != 0 {
					t.Errorf("Expected %v deposits, got %v", 0, len(deposits))
				}
			},
		},
		{
			name:       "goal not found",
			wantStatus: http.StatusNotFound,
			wantErr:    "error finding goal",
			seedDB: func(t *testing.T) (database.User, database.Goal, []database.Deposit) {
				return seedUser(t, "foo-bar-getdeposit3@example.com"), database.Goal{}, nil
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "mismatch user id",
			wantStatus: http.StatusForbidden,
			wantErr:    "mismatch user id",
			seedDB: func(t *testing.T) (database.User, database.Goal, []database.Deposit) {
				user, goal := seedUserWithGoal(t, "foo-bar-getdeposit4@example.com", 1000, time.Now().AddDate(1, 0, 0))

				return user, goal, nil
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(uuid.New(), JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "invalid goal id",
			wantStatus: http.StatusBadRequest,
			wantErr:    "error parsing goal id",
			seedDB: func(t *testing.T) (database.User, database.Goal, []database.Deposit) {
				user, goal := seedUserWithGoal(t, "foo-bar-getdeposit5@example.com", 1000, time.Now().AddDate(1, 0, 0))

				return user, goal, nil
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/goals/abc/deposits", testServer.URL), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(u.ID, JWTSecret, time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
		{
			name:       "authorization header not present",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "authorization header not present",
			seedDB: func(t *testing.T) (database.User, database.Goal, []database.Deposit) {
				user, goal := seedUserWithGoal(t, "foo-bar-getdeposit6@example.com", 1000, time.Now().AddDate(1, 0, 0))

				return user, goal, nil
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
			},
		},
		{
			name:       "expired jwt",
			wantStatus: http.StatusUnauthorized,
			wantErr:    "token is expired",
			seedDB: func(t *testing.T) (database.User, database.Goal, []database.Deposit) {
				user, goal := seedUserWithGoal(t, "foo-bar-getdeposit7@example.com", 1000, time.Now().AddDate(1, 0, 0))

				return user, goal, nil
			},
			setupRequest: func(g database.Goal, t *testing.T) *http.Request {
				req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%v/api/goals/%v/deposits", testServer.URL, g.ID), nil)
				if err != nil {
					t.Fatalf("error creating new request: %v", err)
				}

				return req
			},
			setupHeaders: func(u database.User, r *http.Request) {
				jwt, _ := auth.MakeJWT(uuid.New(), JWTSecret, -time.Hour)
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, goal, seededDeposits := tt.seedDB(t)

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
				return
			}

			if tt.checkDeposits != nil {
				var deposits []goals.Deposit
				decoder := json.NewDecoder(resp.Body)
				err := decoder.Decode(&deposits)
				if err != nil {
					t.Fatalf("failed to decode deposits: %v", err)
				}

				tt.checkDeposits(t, deposits, seededDeposits)
			}
		})
	}
}
