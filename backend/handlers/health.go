package handlers

import (
	"fmt"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

func (a *ApiConfig) CheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	int, err := a.DatabaseQueries.Ping(r.Context())
	if err != nil {
		response.RespondWithError(w, http.StatusServiceUnavailable, fmt.Sprintf("Error pinging database: %v", err))
		return
	}

	response.RespondWithJSON(w, http.StatusOK, int)
}
