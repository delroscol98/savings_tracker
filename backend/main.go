package main

import (
	"context"
	"database/sql"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/delroscol98/savings_tracker/backend/handlers"
	"github.com/delroscol98/savings_tracker/backend/internal/database"
	"github.com/joho/godotenv"

	_ "github.com/lib/pq"
)

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
	api := &handlers.ApiConfig{
		DatabaseQueries: dbQueries,
	}

	// SERVER MULTIPLEXER
	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", http.StripPrefix("/app/", api.MiddlewareMetricInc(http.FileServer(http.Dir(ROOTDIR)))))
	serveMux.Handle("/health", api.MiddlewareLog(http.HandlerFunc(api.CheckHealthHandler)))

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
