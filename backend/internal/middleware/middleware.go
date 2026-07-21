package middleware

import (
	"log"
	"net/http"
)

type HitCounter interface {
	Add(int32) int32
	Load() int32
}

func MetricInc(counter HitCounter, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Increments how many times the server has been hit
		newServerHits := counter.Add(1)
		log.Printf("Hits: %v\n", newServerHits)

		next.ServeHTTP(w, r)
	})
}

func Log(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Logs every request
		log.Printf("%s %s", r.Method, r.URL.Path)

		next.ServeHTTP(w, r)
	})
}
