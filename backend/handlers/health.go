package handlers

import (
	"fmt"
	"net/http"
)

func (a *ApiConfig) CheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	int, err := a.DatabaseQueries.Ping(r.Context())
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, fmt.Sprintf("Error pinging database: %v", err))
		return
	}

	respondWithJSON(w, http.StatusOK, int)
}
