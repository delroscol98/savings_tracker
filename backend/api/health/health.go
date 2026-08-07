package health

import (
	"context"
	"log"
	"net/http"

	"github.com/delroscol98/savings_tracker/backend/internal/response"
)

type Pinger interface {
	Ping(ctx context.Context) (int32, error)
}

type HealthConfig struct {
	Queries Pinger
}

func (h *HealthConfig) CheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	int, err := h.Queries.Ping(r.Context())
	if err != nil {
		log.Print(err)
		response.RespondWithError(w, http.StatusServiceUnavailable, "Error pinging database")
		return
	}

	response.RespondWithJSON(w, http.StatusOK, int)
}
