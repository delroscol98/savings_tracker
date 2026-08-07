package api_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/api/users"
	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
)

func TestCreateUserHandler_Integration(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		fullname   string
		wantStatus int
		wantErr    string
		seedDB     func(*testing.T)
		checkUser  func(t *testing.T, u *users.User)
	}{
		{
			name:       "valid user",
			email:      "test1@example.com",
			password:   "ThisIsATestPassword",
			fullname:   "John Smith",
			wantStatus: http.StatusCreated,
			wantErr:    "",
			seedDB:     func(t *testing.T) {},
			checkUser: func(t *testing.T, u *users.User) {
				if u.Id == uuid.Nil {
					t.Error("User ID should not be zero-value")
				}
				if u.CreatedAt.IsZero() {
					t.Error("User CreatedAt timestamp should not be zero-value")
				}
				if u.UpdatedAt.IsZero() {
					t.Error("User UpdatedAt timestamp should not be zero-value")
				}
			},
		},
		{
			name:       "duplicate email",
			email:      "test2@example.com",
			password:   "ThisIsATestPassword",
			fullname:   "John Smith",
			wantStatus: http.StatusConflict,
			wantErr:    "Email already exists",
			seedDB: func(t *testing.T) {
				hashedPw, err := auth.HashPassword("ThisIsATestPassword")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}

				_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "test2@example.com",
					HashedPassword: hashedPw,
				})
				if err != nil {
					t.Fatalf("Failed to seed new user: %v", err)
				}
			},
			checkUser: nil,
		},
		{
			name:       "empty email",
			email:      "",
			password:   "ThisIsATestPassword",
			fullname:   "John Smith",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			seedDB:     func(t *testing.T) {},
			checkUser:  nil,
		},
		{
			name:       "invalid email",
			email:      "invalidemail",
			password:   "ThisIsATestPassword",
			fullname:   "John Smith",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			seedDB:     func(t *testing.T) {},
			checkUser:  nil,
		},
		{
			name:       "empty password",
			email:      "test3@example.com",
			password:   "",
			fullname:   "John Smith",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			seedDB:     func(t *testing.T) {},
			checkUser:  nil,
		},
		{
			name:       "too short password",
			email:      "test4@example.com",
			password:   "test",
			fullname:   "John Smith",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			seedDB:     func(t *testing.T) {},
			checkUser:  nil,
		},
		{
			name:       "too long password",
			email:      "test5@example.com",
			password:   "ThisPasswordIsLongerThan128CharactersSoThatWeCanTestThatOurSystemHandlesItProperlyWithoutAnyIssuesOrUnexpectedBehaviorWhenCreatingAUserWithThisVeryLongPassword1234567890",
			fullname:   "John Smith",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			seedDB:     func(t *testing.T) {},
			checkUser:  nil,
		},
		{
			name:       "empty full name",
			email:      "test5@example.com",
			password:   "ThisPasswordIsLongerThan128CharactersSoThatWeCanTestThatOurSystemHandlesItProperlyWithoutAnyIssuesOrUnexpectedBehaviorWhenCreatingAUserWithThisVeryLongPassword1234567890",
			fullname:   "",
			wantStatus: http.StatusBadRequest,
			wantErr:    "Invalid parameters to create new user",
			seedDB:     func(t *testing.T) {},
			checkUser:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedDB(t)

			params := strings.NewReader(fmt.Sprintf(`{"email": "%v", "password": "%v", "full_name": "%v"}`, tt.email, tt.password, tt.fullname))
			resp, err := http.Post(testServer.URL+"/api/users", "application/json", params)
			if err != nil {
				t.Fatalf("Error sending post request: %v", err)
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

			if tt.checkUser != nil {
				user := users.User{}
				decoder := json.NewDecoder(resp.Body)
				err := decoder.Decode(&user)
				if err != nil {
					t.Fatalf("Failed to decode user: %v", err)
				}

				if user.Email != tt.email {
					t.Errorf(`
Expected email: %v,
Actual email:   %v
`, tt.email, user.Email)
				}

				tt.checkUser(t, &user)
			}
		})
	}
}
