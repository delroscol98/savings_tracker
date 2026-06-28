package handlers_test

import (
	"bytes"
	"database/sql"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

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
