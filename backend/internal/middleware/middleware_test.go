package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/delroscol98/savings_tracker/backend/internal/middleware"
)

var apiHits atomic.Int32

func TestMiddlewareMetricInc(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := middleware.MetricInc(&apiHits, next)
	w := httptest.NewRecorder()
	r := httptest.NewRequest(http.MethodGet, "/app/", nil)
	handler.ServeHTTP(w, r)

	if w.Code != http.StatusTeapot {
		t.Errorf(`
Expecting status code: %v
Actual status code:    %v
`, http.StatusTeapot, w.Code)
	}

	hits := apiHits.Load()
	if hits != 1 {
		t.Errorf(`
Expecting FileserverHits: 1
Actual FileserverHits:    %v
`, hits)
	}
}
