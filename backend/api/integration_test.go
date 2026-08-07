package api_test

import (
	"bytes"
	"database/sql"
	"log"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/delroscol98/savings_tracker/backend/api/goals"
	"github.com/delroscol98/savings_tracker/backend/api/health"
	"github.com/delroscol98/savings_tracker/backend/api/users"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
	"github.com/delroscol98/savings_tracker/backend/internal/router"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

var (
	testServer      *httptest.Server
	dbQueries       *database.Queries
	testRateLimiter *ratelimit.RateLimiter
	mockSender      *MockEmailSender
	JWTSecret       string
	baseURL         string
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

	migration005, err := os.ReadFile("../sql/schema/005_create_goals_and_deposits.sql")
	if err != nil {
		log.Fatalf("Error reading schema: %v", err)
	}
	migration005Split := bytes.Split(migration005, []byte("\n-- +goose Down\n"))
	upMigration005 := migration005Split[0]
	downMigration005 := migration005Split[1]
	_, err = db.Exec(string(upMigration005))
	if err != nil {
		log.Fatalf("Error executing up migration: %v", err)
	}

	testRateLimiter = ratelimit.New(5, 15*time.Minute)
	dbQueries = database.New(db)

	mockSender = &MockEmailSender{}

	JWTSecret = "secret"
	os.Setenv("JWT_SECRET", JWTSecret)

	baseURL = "https://localhost:8080"
	usersAPI := &users.UsersConfig{
		Queries:     dbQueries,
		Database:    db,
		RateLimiter: testRateLimiter,
		EmailSender: mockSender,
		JWTSecret:   JWTSecret,
		BaseURL:     baseURL,
	}
	goalsAPI := &goals.GoalsConfig{
		Queries: dbQueries,
	}
	healthApi := &health.HealthConfig{
		Queries: dbQueries,
	}

	testServer = httptest.NewServer(router.NewRouter(router.Dependencies{
		Users:  usersAPI,
		Goals:  goalsAPI,
		Health: healthApi,
	}))

	code := m.Run()
	_, err = db.Exec(string(downMigration005))
	if err != nil {
		log.Fatalf("Error executing down migration: %v", err)
	}

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
