package goals_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/api/goals"
	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
)

func TestCreateDepositHandler(t *testing.T) {
	userId := uuid.New()
	tokenSecret = "secret"
	expiresIn := time.Hour

	token, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

	goalId := uuid.New()

	tests := []struct {
		name         string
		userId       uuid.UUID
		goalId       uuid.UUID
		amount       int
		note         string
		wantStatus   int
		wantErr      string
		setupHeaders func(*http.Request)
		seedDB       func(*mockDB)
	}{
		{
			name:       "valid deposit",
			userId:     userId,
			goalId:     goalId,
			amount:     100,
			note:       "test deposit",
			wantStatus: http.StatusCreated,
			wantErr:    "",
			setupHeaders: func(r *http.Request) {
				r.Header.Set("Content-Type", "application/json")
				r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", token))
			},
			seedDB: func(md *mockDB) {
				hash, _ := auth.HashPassword("ThisIsATestPassword")
				user := database.User{
					ID:             userId,
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					Email:          "test@example.com",
					HashedPassword: hash,
					FullName:       "John Smith",
				}

				md.Users[userId.String()] = user

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDB{
				Goals: make(map[string]database.Goal),
				Users: make(map[string]database.User),
			}
			tt.seedDB(md)

			api := goals.GoalsConfig{Queries: md}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/goals/%v", goalId), strings.NewReader(fmt.Sprintf(`{"amount": %v, "note": "%v"}`, tt.amount, tt.note)))
			r.SetPathValue("goalId", tt.goalId.String())
			tt.setupHeaders(r)
			api.CreateDepositHandler(w, withUserContext(r, userId))

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
