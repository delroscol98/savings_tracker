package handlers_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/delroscol98/savings_tracker/backend/internal/auth"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var (
	testServer      *httptest.Server
	dbQueries       *database.Queries
	testRateLimiter *ratelimit.RateLimiter
	mockSender      *MockEmailSender
)

func TestMain(m *testing.M) {
	// DATABASE SETUP
	_ = godotenv.Load("../.env")

	dbURL := os.Getenv("DB_URL_TEST")
	if dbURL == "" {
		log.Fatalf("DB_URL_TEST not set")
	}

	db, _ := sql.Open("postgres", dbURL)
	err := db.Ping()
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// MIGRATIONS
	migration001, err := os.ReadFile("../sql/schema/001_create_users.sql")
	if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	migration001Split := bytes.Split(migration001, []byte("\n-- +goose Down\n"))
	upMigration001 := migration001Split[0]
	downMigration001 := migration001Split[1]
	_, err = db.Exec(string(upMigration001))
	if err != nil {
		log.Fatalf("Error executing up migration: %v", err)
	}

	migration002, err := os.ReadFile("../sql/schema/002_add_hashed_passwords_to_users.sql")
	if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	migration002Split := bytes.Split(migration002, []byte("\n-- +goose Down\n"))
	upMigration002 := migration002Split[0]
	downMigration002 := migration002Split[1]
	_, err = db.Exec(string(upMigration002))
	if err != nil {
		log.Fatalf("Error executing up migration: %v", err)
	}

	migration003, err := os.ReadFile("../sql/schema/003_add_fullname_to_users.sql")
	if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	migration003Split := bytes.Split(migration003, []byte("\n-- +goose Down\n"))
	upMigration003 := migration003Split[0]
	downMigration003 := migration003Split[1]
	_, err = db.Exec(string(upMigration003))
	if err != nil {
		log.Fatalf("Error executing up migration: %v", err)
	}

	migration004, err := os.ReadFile("../sql/schema/004_create_password_reset_tokens.sql")
	if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	migration004Split := bytes.Split(migration004, []byte("\n-- +goose Down\n"))
	upMigration004 := migration004Split[0]
	downMigration004 := migration004Split[1]
	_, err = db.Exec(string(upMigration004))
	if err != nil {
		log.Fatalf("Error executing up migration: %v", err)
	}

	testRateLimiter = ratelimit.New(5, 15*time.Minute)
	dbQueries = database.New(db)

	mockSender = &MockEmailSender{}
	api := &handlers.ApiConfig{
		DatabaseQueries: dbQueries,
		Db:              db,
		RateLimiter:     testRateLimiter,
		EmailSender:     mockSender,
	}

	// SERVER MULTIPLEXER
	serveMux := http.NewServeMux()
	serveMux.Handle("GET /health", api.MiddlewareLog(http.HandlerFunc(api.CheckHealthHandler)))
	serveMux.Handle("POST /api/users", api.MiddlewareLog(http.HandlerFunc(api.CreateUserHandler)))
	serveMux.Handle("POST /api/login", api.MiddlewareLog(http.HandlerFunc(api.LoginUserHandler)))
	serveMux.Handle("POST /api/forgot-password", api.MiddlewareLog(http.HandlerFunc(api.RequestPasswordResetHandler)))
	serveMux.Handle("POST /api/reset-password", api.MiddlewareLog(http.HandlerFunc(api.ResetPasswordHandler)))

	testServer = httptest.NewServer(serveMux)

	code := m.Run()
	_, err = db.Exec(string(downMigration004))
	if err != nil {
		log.Fatalf("Error executing down migration: %v", err)
	}

	_, err = db.Exec(string(downMigration003))
	if err != nil {
		log.Fatalf("Error executing down migration: %v", err)
	}

	_, err = db.Exec(string(downMigration002))
	if err != nil {
		log.Fatalf("Error executing down migration: %v", err)
	}

	_, err = db.Exec(string(downMigration001))
	if err != nil {
		log.Fatalf("Error executing down migration: %v", err)
	}
	testServer.Close()
	db.Close()
	os.Exit(code)
}

func TestCreateUserHandler_Integration(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		fullname   string
		wantStatus int
		wantErr    string
		seedDB     func(*testing.T)
		checkUser  func(*testing.T, *handlers.User)
	}{
		{
			name:       "valid user",
			email:      "test1@example.com",
			password:   "ThisIsATestPassword",
			fullname:   "John Smith",
			wantStatus: http.StatusCreated,
			wantErr:    "",
			seedDB:     func(t *testing.T) {},
			checkUser: func(t *testing.T, u *handlers.User) {
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
				user := handlers.User{}
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

func TestLoginHandler_Integration(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		password   string
		wantStatus int
		wantErr    string
		seedDB     func(*testing.T)
	}{
		{
			name:       "valid login",
			email:      "foo-1@example.com",
			password:   "AnotherTestPassword",
			wantStatus: http.StatusOK,
			wantErr:    "",
			seedDB: func(t *testing.T) {
				hashedPw, err := auth.HashPassword("AnotherTestPassword")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}

				_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-1@example.com",
					HashedPassword: hashedPw,
				})
				if err != nil {
					t.Fatalf("Failed to seed new user: %v", err)
				}
			},
		},
		{
			name:       "incorrect email",
			email:      "foo-3@example.com",
			password:   "AnotherTestPassword",
			wantStatus: http.StatusBadRequest,
			wantErr:    "User not found",
			seedDB: func(t *testing.T) {
				hashedPw, err := auth.HashPassword("ThisIsATestPassword")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}

				_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-2@example.com",
					HashedPassword: hashedPw,
				})
				if err != nil {
					t.Fatalf("Failed to seed new user: %v", err)
				}
			},
		},
		{
			name:       "incorrect password",
			email:      "foo-4@example.com",
			password:   "AnotherTestPassword",
			wantStatus: http.StatusForbidden,
			wantErr:    "Incorrect email or password",
			seedDB: func(t *testing.T) {
				hashedPw, err := auth.HashPassword("ThisIsATestPassword")
				if err != nil {
					t.Fatalf("Failed to hash password: %v", err)
				}

				_, err = dbQueries.CreateUser(context.Background(), database.CreateUserParams{
					Email:          "foo-4@example.com",
					HashedPassword: hashedPw,
				})
				if err != nil {
					t.Fatalf("Failed to seed new user: %v", err)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.seedDB(t)

			params := strings.NewReader(fmt.Sprintf(`{"email": "%v", "password": "%v"}`, tt.email, tt.password))
			resp, err := http.Post(testServer.URL+"/api/login", "application/json", params)
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
		})
	}
}

func extractToken(t *testing.T, logOutput string) string {
	t.Helper()
	re := regexp.MustCompile(`token=([a-f0-9]+)`)
	matches := re.FindStringSubmatch(logOutput)
	if len(matches) < 2 {
		t.Fatalf("Token not found in log output: %v", logOutput)
	}
	return matches[1]
}

func TestFullPasswordResetFlow_Integration(t *testing.T) {
	testRateLimiter.Reset()

	// Create user
	createUserBody := strings.NewReader(`
{
	"email": "johnsmith@testexample.com",
	"password": "AnotherTestPassword",
	"full_name": "John Smith"
}
	`)
	resp, err := http.Post(testServer.URL+"/api/users", "application/json", createUserBody)
	if err != nil {
		t.Fatalf("Error creating user in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf(`
Creating a new user:

Expected status code: %v
Actual status code:   %v
`, http.StatusCreated, resp.StatusCode)
	}

	// Forgot password
	buf := bytes.Buffer{}
	log.SetOutput(&buf)
	defer log.SetOutput(os.Stderr)

	forgotPasswordBody := strings.NewReader(`
{
	"email": "johnsmith@testexample.com"
}
	`)
	resp, err = http.Post(testServer.URL+"/api/forgot-password", "application/json", forgotPasswordBody)
	if err != nil {
		t.Fatalf("Error creating reset password link in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`
Creating a reset password link:

Expected status code: %v
Actual status code:   %v
`, http.StatusOK, resp.StatusCode)
	}

	if len(mockSender.Sent) != 1 {
		t.Fatalf(`
Expected number of sent emails: 1
Actual number of sent emails:   %v
`, len(mockSender.Sent))
	}

	if mockSender.Sent[0].To != "johnsmith@testexample.com" {
		t.Fatalf(`
Expected To email: johnsmith@testexample.com
Actual to email:   %v
`, mockSender.Sent[0].To)
	}

	token := extractToken(t, buf.String())

	// Reset Password
	resetPasswordBody := strings.NewReader(fmt.Sprintf(`
{
	"token": "%v",
	"password": "DifferentTestPassword"
}
	`, token))

	resp, err = http.Post(testServer.URL+"/api/reset-password", "application/json", resetPasswordBody)
	if err != nil {
		t.Fatalf("Error resetting password in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`
Resetting password:

Expected status code: %v
Actual status code:   %v
`, http.StatusOK, resp.StatusCode)
	}

	// Login with new password "DifferentTestPassword"
	loginBody := strings.NewReader(`
{
	"email": "johnsmith@testexample.com",
	"password": "DifferentTestPassword"
}
`)

	resp, err = http.Post(testServer.URL+"/api/login", "application/json", loginBody)
	if err != nil {
		t.Fatalf("Error logging in with new password in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf(`
Logging in:

Expected status code: %v,
Actual status code:   %v,
`, http.StatusOK, resp.StatusCode)
	}

	// Login with old password "AnotherTestPassword"
	loginBody = strings.NewReader(`
{
	"email": "johnsmith@testexample.com",
	"password": "AnotherTestPassword"
}
`)

	resp, err = http.Post(testServer.URL+"/api/login", "application/json", loginBody)
	if err != nil {
		t.Fatalf("Error logging in with old password in integration test: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Errorf(`
Logging in:

Expected status code: %v,
Actual status code:   %v,
`, http.StatusForbidden, resp.StatusCode)
	}
}
