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
