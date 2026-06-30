package handlers_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var testServer *httptest.Server

func TestMain(m *testing.M) {
	// DATABASE SETUP
	err := godotenv.Load("../.env")
	if err != nil {
		log.Fatalf("Error connecting to env file: %s", err)
	}
	dbURL := os.Getenv("DB_URL_TEST")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}

	// MIGRATIONS
	migration, err := os.ReadFile("../sql/schema/001_users.sql")
	if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	migrationSplit := bytes.Split(migration, []byte("\n-- +goose Down\n"))
	upMigration := migrationSplit[0]
	downMigration := migrationSplit[1]
	_, err = db.Exec(string(upMigration))
	if err != nil {
		log.Fatalf("Error executing up migration: %v", err)
	}

	dbQueries := database.New(db)
	api := &handlers.ApiConfig{
		DatabaseQueries: dbQueries,
	}

	// SERVER MULTIPLEXER
	serveMux := http.NewServeMux()
	serveMux.Handle("GET /health", api.MiddlewareLog(http.HandlerFunc(api.CheckHealthHandler)))
	serveMux.Handle("POST /api/users", api.MiddlewareLog(http.HandlerFunc(api.CreateUserHandler)))

	testServer = httptest.NewServer(serveMux)

	code := m.Run()
	_, err = db.Exec(string(downMigration))
	if err != nil {
		log.Fatalf("Error executing down migration: %v", err)
	}
	testServer.Close()
	db.Close()
	os.Exit(code)
}

func TestCreateUserHandler_Valid_Integration(t *testing.T) {
	body := strings.NewReader(`{"email": "test@example.com"}`)
	resp, err := http.Post(testServer.URL+"/api/users", "application/json", body)
	if err != nil {
		t.Error("Error creating user")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf(`
Expected status code: 201
Actual status code:   %v`, resp.StatusCode)
	}

	user := handlers.User{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&user)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v`, err)
	}

	if user.Email != "test@example.com" {
		t.Errorf(`
Expected email: test@example.com
Actual email:   %v`, user.Email)
	}

	if user.Id == uuid.Nil {
		t.Error("user ID should NOT be UUID zero-value")
	}

	if user.CreatedAt.IsZero() {
		t.Error("user created_at should NOT be timestamp zero-value")
	}

	if user.UpdatedAt.IsZero() {
		t.Error("user updated_at should NOT be timestamp zero-value")
	}
}

func TestCreateUserHandler_DuplicateEmail_Integration(t *testing.T) {
	body := strings.NewReader(`{"email": "test@example.com"}`)
	resp, err := http.Post(testServer.URL+"/api/users", "application/json", body)
	if err != nil {
		t.Error("Error creating user")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf(`
Expected status code: 409
Actual status code:   %v`, resp.StatusCode)
	}

	errorBody := handlers.ErrorBody{}
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&errorBody)
	if err != nil {
		t.Errorf(`
Expected error: nil
Actual error:   %v`, err)
	}

	if errorBody.Error != "Email already exists" {
		t.Errorf(`
Expected error: Email already exists
Actual error:   %v`, errorBody.Error)
	}
}
