package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delroscol98/savings_tracker/backend/api/health"
)

type mockHealthDB struct {
	pingErr error
}

func (m *mockHealthDB) Ping(ctx context.Context) (int32, error) {
	return 1, m.pingErr
}

func TestCheckHealthHandler_DBHealthy(t *testing.T) {
	api := health.HealthConfig{
		Queries: &mockHealthDB{},
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
	api := health.HealthConfig{
		Queries: &mockHealthDB{pingErr: errors.New("database down")},
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

	if body["error"] != "Error pinging database" {
		t.Errorf(`
Expected error message: Error pinging database
Actual error message:   %v
`, body["error"])
	}
}
