package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/delroscol98/savings_tracker/backend/api/goals"
	"github.com/delroscol98/savings_tracker/backend/api/health"
	"github.com/delroscol98/savings_tracker/backend/api/users"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
	"github.com/delroscol98/savings_tracker/backend/internal/router"
	"github.com/joho/godotenv"
	"github.com/resend/resend-go/v3"

	_ "github.com/lib/pq"
)

var serverHits atomic.Int32

func main() {
	// CONSTANTS
	const PORT = ":8080"
	const ROOTDIR = "./"

	// DATABASE
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error connecting to env file: %s", err)
	}
	dbURL := os.Getenv("DB_URL")
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	dbQueries := database.New(db)

	// Resend
	resendApiKey := os.Getenv("RESEND_API_KEY")
	fromEmail := os.Getenv("FROM_EMAIL")
	resendSender := users.ResendSender{
		Client: resend.NewClient(resendApiKey),
		From:   fromEmail,
	}

	// JWT
	secret := os.Getenv("JWT_SECRET")

	// BASE URL
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		log.Fatal("Base URL must be set")
	}

	// API setup
	usersAPI := &users.UsersConfig{
		Queries:                  dbQueries,
		PasswordResetRateLimiter: ratelimit.New(5, 15*time.Minute),
		LoginRateLimiter:         ratelimit.New(10, 15*time.Minute),
		LoginThrottler:           ratelimit.NewLoginThrottler(5, 15*time.Minute),
		Database:                 db,
		EmailSender:              &resendSender,
		JWTSecret:                secret,
		BaseURL:                  baseURL,
	}
	goalsAPI := &goals.GoalsConfig{
		Queries: dbQueries,
	}
	healthApi := &health.HealthConfig{
		Queries: dbQueries,
	}

	// SERVER MULTIPLEXER
	serveMux := router.NewRouter(router.Dependencies{
		Users:  usersAPI,
		Goals:  goalsAPI,
		Health: healthApi,
	})
	serveMux.Handle("GET /app", http.StripPrefix("/app", middleware.MetricInc(&serverHits, http.FileServer(http.Dir(ROOTDIR)))))

	// START THE SERVER
	server := http.Server{
		Handler:      serveMux,
		Addr:         PORT,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	log.Printf("Serving files from %s on port %s\n", ROOTDIR, PORT)
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		server.Shutdown(ctx) // wait for active requests to finish
	}()
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
