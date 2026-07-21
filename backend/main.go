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

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/delroscol98/savings_tracker/backend/handlers/auth"
	"github.com/delroscol98/savings_tracker/backend/handlers/health"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
	"github.com/delroscol98/savings_tracker/backend/internal/ratelimit"
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
	resendSender := handlers.ResendSender{
		Client: resend.NewClient(resendApiKey),
		From:   fromEmail,
	}

	// API setup
	authApi := &auth.AuthConfig{
		Queries:     dbQueries,
		RateLimiter: ratelimit.New(5, 15*time.Minute),
		Database:    db,
		EmailSender: &resendSender,
	}
	healthApi := &health.HealthConfig{
		Queries: dbQueries,
	}

	// SERVER MULTIPLEXER
	serveMux := http.NewServeMux()
	serveMux.Handle("GET /app", http.StripPrefix("/app", middleware.MetricInc(&serverHits, http.FileServer(http.Dir(ROOTDIR)))))
	serveMux.Handle("GET /health", middleware.Log(http.HandlerFunc(healthApi.CheckHealthHandler)))
	serveMux.Handle("POST /api/users", middleware.Log(http.HandlerFunc(authApi.CreateUserHandler)))
	serveMux.Handle("POST /api/login", middleware.Log(http.HandlerFunc(authApi.LoginUserHandler)))
	serveMux.Handle("POST /api/forgot-password", middleware.Log(http.HandlerFunc(authApi.RequestPasswordResetHandler)))
	serveMux.Handle("POST /api/reset-password", middleware.Log(http.HandlerFunc(authApi.ResetPasswordHandler)))

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
