package handlers

import (
	"encoding/json"
	"net/http"
)

func (s *ApiConfig) CheckHealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}
