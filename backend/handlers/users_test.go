package handlers_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/google/uuid"
)

func TestCreateUserHandler(t *testing.T) {
	rawPw := "ThisIsATestPassword"

	tests := []struct {
		name       string
		body       io.Reader
		wantStatus int
		wantErr    string
		wantEmail  string
		setupMock  func(*mockDB)
		checkUser  func(*testing.T, handlers.User)
	}{
		{
			name:       "valid user",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "` + rawPw + `"}`),
			wantStatus: http.StatusCreated,
			wantEmail:  "test@example.com",
			setupMock:  func(md *mockDB) {},
			checkUser: func(t *testing.T, u handlers.User) {
				if u.Id == uuid.Nil {
					t.Error("user ID should not be zero-value")
				}
				if u.CreatedAt.IsZero() {
					t.Error("created_at should not be zero-value")
				}
				if u.UpdatedAt.IsZero() {
					t.Error("updated_at should not be zero-value")
				}
			},
		},
		{
			name:       "empty email",
			body:       strings.NewReader(`{"email": "", "password": "` + rawPw + `"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters for user action",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "invalid email",
			body:       strings.NewReader(`{"email": "ThisIsAnInvalidEmail", "password": "` + rawPw + `"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters for user action",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "empty password",
			body:       strings.NewReader(`{"email": "test@example.com", "password": ""}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters for user action",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "too short password",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "test"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters for user action",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "too long password",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "ThisPasswordIsLongerThan128CharactersSoThatWeCanTestThatOurSystemHandlesItProperlyWithoutAnyIssuesOrUnexpectedBehaviorWhenCreatingAUserWithThisVeryLongPassword1234567890"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters for user action",
			setupMock:  func(md *mockDB) {},
		},
		{
			name: "duplicate email",
			body: strings.NewReader(`{"email": "test@example.com", "password": "` + rawPw + `"}`),
			setupMock: func(md *mockDB) {
				md.users["test@example.com"] = database.User{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
					Email:     "test@example.com",
				}
			},
			wantStatus: http.StatusConflict,
			wantErr:    "Email already exists",
		},
		{
			name:       "empty body",
			body:       nil,
			wantStatus: http.StatusBadRequest,
			wantErr:    "Error decoding body",
			setupMock:  func(md *mockDB) {},
		},
		{
			name: "db error",
			body: strings.NewReader(`{"email": "test@example.com", "password": "` + rawPw + `"}`),
			setupMock: func(md *mockDB) {
				md.CreateUserErr = errors.New("connection refused")
			},
			wantStatus: http.StatusBadRequest,
			wantErr:    "Error creating user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := &mockDB{users: make(map[string]database.User)}
			tt.setupMock(md)

			api := handlers.ApiConfig{DatabaseQueries: md}
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/api/users", tt.body)
			r.Header.Set("Content-Type", "application/json")
			api.CreateUserHandler(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("want status %d, got %d", tt.wantStatus, w.Code)
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

			if tt.checkUser != nil {
				var user handlers.User
				if err := json.NewDecoder(w.Body).Decode(&user); err != nil {
					t.Fatalf("failed to decode user: %v", err)
				}
				if user.Email != tt.wantEmail {
					t.Errorf("want email %q, got %q", tt.wantEmail, user.Email)
				}
				tt.checkUser(t, user)
			}
		})
	}
}
