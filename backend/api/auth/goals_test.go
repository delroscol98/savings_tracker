package auth_test

import (
	"encoding/json"
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

func TestCreateGoalHandler(t *testing.T) {
	userId := uuid.New()
	tokenSecret := "secret"
	expiresIn := time.Hour

	jwt, _ := auth.MakeJWT(userId, tokenSecret, expiresIn)

	tests := []struct {
		name       string
		body       io.Reader
		wantStatus int
		wantErr    string
		setupMock  func(*mockDB)
	}{
		{
			name:       "valid goal",
			body:       strings.NewReader(fmt.Sprintf(`{"target": 1000, "deadline": "%s", "user_id": "%v"}`, time.Now().Add(expiresIn).Format(time.RFC3339), userId)),
			wantStatus: http.StatusCreated,
			wantErr:    "",
			setupMock:  func(md *mockDB) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDB{
				users: make(map[string]database.User),
				Goals: make(map[string]database.Goal),
			}
			tt.setupMock(md)

			api := auth.AuthConfig{Queries: md}
			api.JWTSecret = tokenSecret
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/goals", tt.body)
			r.Header.Set("Content-Type", "application/json")
			r.Header.Set("Authorization", fmt.Sprintf("Bearer %v", jwt))
			api.CreateGoalHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf(`
Expected status code: %v
Actual status code:   %v
`, tt.wantStatus, w.Code)
			}

			t.Log(w.Body)

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
