package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

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
	serveMux.HandleFunc("/health", api.MiddlewareLog(api.CheckHealthHandler))

	// START THE SERVER
	server := http.Server{
		Handler: serveMux,
		Addr:    PORT,
	}
	log.Printf("Serving files from %s on port %s\n", ROOTDIR, PORT)
	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
