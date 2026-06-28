package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delroscol98/savings_tracker/backend/handlers"
)

func TestCheckHealthHandler_DBHealthy(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	api.CheckHealthHandler(w, r)

	if w.Code != http.StatusOK {
		t.Errorf(`
Expected status code: 200
Actual status code:   %v
`, w.Code)
	}

	if w.Body.String() != "1" {
		t.Errorf(`
Expected body: "1"
Actual body:   %v
`, w.Body.String())
	}
}

func TestCheckHealthHandler_DBUnhealthy(t *testing.T) {
	api := handlers.ApiConfig{
		DatabaseQueries: &mockDB{pingErr: errors.New("database down")},
	}
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/health", nil)
	api.CheckHealthHandler(w, r)

	if w.Code != http.StatusServiceUnavailable {
		t.Errorf(`
Expected status code: 503
Actual status code:   %v
`, w.Code)
	}

	var body map[string]interface{}
	json.NewDecoder(w.Body).Decode(&body)

	if body["error"] != "Error pinging database: database down" {
		t.Errorf(`
Expected error message: Error pinging database: database down
Actual error message:   %v
`, body["error"])
	}
}
