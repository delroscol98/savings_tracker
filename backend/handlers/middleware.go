package handlers

import (
	"log"
	"net/http"
)

func (a *ApiConfig) MiddlewareMetricInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increments how many times the server has been hit
		newServerHits := a.FileserverHits.Add(1)
		log.Printf("Hits: %v\n", newServerHits)

		next.ServeHTTP(w, r)
	})
}

func (a *ApiConfig) MiddlewareLog(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logs every request
		log.Printf("%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
