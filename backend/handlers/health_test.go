package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delroscol98/savings_tracker/backend/handlers"
)

// TestCheckHealthHandler determines of the API works
// Should automatically return a 200 response
func TestCheckHealthHandler(t *testing.T) {
	apiConfig := handlers.ApiConfig{}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	apiConfig.CheckHealthHandler(w, r)

	if w.Code != http.StatusOK {
		t.Error("API health check does not work")
	}
}
