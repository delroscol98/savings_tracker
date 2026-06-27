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

	api := &handlers.ApiConfig{}

	// Serves static files
	serveMux.Handle("/app/", http.StripPrefix("/app/", api.MiddlewareMetricInc(http.FileServer(http.Dir(ROOTDIR)))))

	// Serves API endpoints
	serveMux.HandleFunc("/health", api.MiddlewareLog(api.CheckHealthHandler))

	server := http.Server{
		Handler: serveMux,
		Addr:    PORT,
	}

	log.Printf("Serving files from %s on port %s\n", ROOTDIR, PORT)
	err := server.ListenAndServe()
	if err != nil {
		log.Fatalf("Error starting server: %s", err)
	}
}
