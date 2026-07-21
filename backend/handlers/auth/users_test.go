package auth_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/handlers/auth"
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
		checkUser  func(*testing.T, auth.User)
	}{
		{
			name:       "valid user",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "` + rawPw + `", "full_name": "John Smith"}`),
			wantStatus: http.StatusCreated,
			wantEmail:  "test@example.com",
			setupMock:  func(md *mockDB) {},
			checkUser: func(t *testing.T, u auth.User) {
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
			body:       strings.NewReader(`{"email": "", "password": "` + rawPw + `", "full_name": "John Smith"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "invalid email",
			body:       strings.NewReader(`{"email": "ThisIsAnInvalidEmail", "password": "` + rawPw + `", "full_name": "John Smith"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "empty password",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "", "full_name": "John Smith"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "too short password",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "test", "full_name": "John Smith"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			setupMock:  func(md *mockDB) {},
		},
		{
			name:       "too long password",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "ThisPasswordIsLongerThan128CharactersSoThatWeCanTestThatOurSystemHandlesItProperlyWithoutAnyIssuesOrUnexpectedBehaviorWhenCreatingAUserWithThisVeryLongPassword1234567890", "full_name": "John Smith"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			setupMock:  func(md *mockDB) {},
		},
		{
			name: "duplicate email",
			body: strings.NewReader(`{"email": "test@example.com", "password": "` + rawPw + `", "full_name": "John Smith"}`),
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
			name:       "empty full name",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "` + rawPw + `", "full_name": ""}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			setupMock:  func(md *mockDB) {},
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
			body: strings.NewReader(`{"email": "test@example.com", "password": "` + rawPw + `", "full_name": "John Smith"}`),
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

			api := auth.AuthConfig{Queries: md}
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
				var user auth.User
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

func TestLoginHandler(t *testing.T) {
	// Cases:
	// 1. valid login
	// 2. incorrect email
	// 3. incorrect password
	// 4. database error

	tests := []struct {
		name       string
		body       io.Reader
		wantStatus int
		wantErr    string
		setupMock  func(*mockDB)
	}{
		{
			name:       "valid login",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "ThisIsATestPassword"}`),
			wantStatus: http.StatusOK,
			wantErr:    "",
			setupMock: func(md *mockDB) {
				password := "ThisIsATestPassword"
				hash, _ := auth.HashPassword(password)

				md.users["test@example.com"] = database.User{
					ID:             uuid.New(),
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					Email:          "test@example.com",
					HashedPassword: hash,
				}
			},
		},
		{
			name:       "incorrect email",
			body:       strings.NewReader(`{"email": "wrong@example.com", "password": "ThisIsATestPassword"}`),
			wantStatus: http.StatusBadRequest,
			wantErr:    "User not found",
			setupMock: func(md *mockDB) {
				password := "ThisIsATestPassword"
				hash, _ := auth.HashPassword(password)

				md.users["test@example.com"] = database.User{
					ID:             uuid.New(),
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					Email:          "test@example.com",
					HashedPassword: hash,
				}
			},
		},
		{
			name:       "incorrect password",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "WrongPassword"}`),
			wantStatus: http.StatusForbidden,
			wantErr:    "Incorrect email or password",
			setupMock: func(md *mockDB) {
				password := "ThisIsATestPassword"
				hash, _ := auth.HashPassword(password)

				md.users["test@example.com"] = database.User{
					ID:             uuid.New(),
					CreatedAt:      time.Now(),
					UpdatedAt:      time.Now(),
					Email:          "test@example.com",
					HashedPassword: hash,
				}
			},
		},
		{
			name:       "database error",
			body:       strings.NewReader(`{"email": "test@example.com", "password": "WrongPassword"}`),
			wantStatus: http.StatusInternalServerError,
			wantErr:    "Unexpected database failure",
			setupMock: func(md *mockDB) {
				md.LoginErr = errors.New("Unexpected database failure")
			},
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				md := &mockDB{users: make(map[string]database.User)}
				tt.setupMock(md)

				api := auth.AuthConfig{Queries: md}
				w := httptest.NewRecorder()
				r := httptest.NewRequest(http.MethodPost, "/api/login", tt.body)
				r.Header.Set("Content-Type", "application/json")
				api.LoginUserHandler(w, r)

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
						t.Errorf(`
Expected error string: %v
Actual error string:   %v
`, tt.wantErr, body["error"])
					}
				}
			},
		)
	}
}
