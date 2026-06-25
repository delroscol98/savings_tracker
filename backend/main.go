package main

import (
	"log"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/handlers"
)

func main() {
	// CONSTANTS
	const PORT = ":8080"
	const ROOTDIR = "./"

	serveMux := http.NewServeMux()

	apiConfig := handlers.ApiConfig{}

	// Serves static files
	serveMux.Handle("/app/", http.StripPrefix("/app/", http.FileServer(http.Dir(ROOTDIR))))
	// serveMux.HandleFunc("/health", apiConfig.CheckHealthHandler)

	server := http.Server{
		Handler: serveMux,
		Addr:    PORT,
	}
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
