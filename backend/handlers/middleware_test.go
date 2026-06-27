package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/delroscol98/savings_tracker/backend/handlers"
)

func TestMiddlewareMetricInc(t *testing.T) {
	api := handlers.ApiConfig{}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := api.MiddlewareMetricInc(next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/app/", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusTeapot {
		t.Errorf(`
Expecting status code: %v
Actual status code:    %v
`, http.StatusTeapot, w.Code)
	}

	hits := api.FileserverHits.Load()
	if hits != 1 {
		t.Errorf(`
Expecting FileserverHits: 1
Actual FileserverHits:    %v
`, hits)
	}
}
