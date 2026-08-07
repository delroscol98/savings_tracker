package api_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func TestLoginRateLimit_Integration(t *testing.T) {
	testLoginRateLimiter.Reset()
	testLoginThrottler.Reset()

	const attempts = 11
	for i := 0; i < attempts; i++ {
		body := strings.NewReader(fmt.Sprintf(`{"email": "rate-limit-%d@example.com", "password": "WrongPassword"}`, i))
		resp, err := http.Post(testServer.URL+"/api/login", "application/json", body)
		if err != nil {
			t.Fatalf("Error sending login request: %v", err)
		}
		resp.Body.Close()

		want := http.StatusForbidden
		if i == attempts-1 {
			want = http.StatusTooManyRequests
		}
		if resp.StatusCode != want {
			t.Errorf("attempt %d: want status %d, got %d", i+1, want, resp.StatusCode)
		}
	}
}

func TestLoginRateLimitByClientIP_Integration(t *testing.T) {
	testLoginRateLimiter.Reset()
	testLoginThrottler.Reset()

	attempt := func(xff string, i int) int {
		body := strings.NewReader(fmt.Sprintf(`{"email": "xff-%s-%d@example.com", "password": "WrongPassword"}`, strings.ReplaceAll(xff, ".", "-"), i))
		req, err := http.NewRequest(http.MethodPost, testServer.URL+"/api/login", body)
		if err != nil {
			t.Fatalf("Error building login request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", xff)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("Error sending login request: %v", err)
		}
		defer resp.Body.Close()
		return resp.StatusCode
	}

	const attempts = 11
	for i := 0; i < attempts; i++ {
		got := attempt("1.2.3.4", i)
		want := http.StatusForbidden
		if i == attempts-1 {
			want = http.StatusTooManyRequests
		}
		if got != want {
			t.Errorf("xff ip 1.2.3.4 attempt %d: want status %d, got %d", i+1, want, got)
		}
	}

	for i := 0; i < 3; i++ {
		if got := attempt("5.6.7.8", i); got != http.StatusForbidden {
			t.Errorf("xff ip 5.6.7.8 attempt %d: want status %d, got %d", i+1, http.StatusForbidden, got)
		}
	}
}

func TestLoginLockout_Integration(t *testing.T) {
	testLoginRateLimiter.Reset()
	testLoginThrottler.Reset()

	seedUserWithPassword(t, "lockout@example.com", "AnotherTestPassword")

	const attempts = 6
	for i := 0; i < attempts; i++ {
		body := strings.NewReader(`{"email": "lockout@example.com", "password": "WrongPassword"}`)
		resp, err := http.Post(testServer.URL+"/api/login", "application/json", body)
		if err != nil {
			t.Fatalf("Error sending login request: %v", err)
		}
		resp.Body.Close()

		want := http.StatusForbidden
		if i == attempts-1 {
			want = http.StatusTooManyRequests
		}
		if resp.StatusCode != want {
			t.Errorf("attempt %d: want status %d, got %d", i+1, want, resp.StatusCode)
		}
	}
}

func TestPasswordResetRateLimit_Integration(t *testing.T) {
	testPasswordResetRateLimiter.Reset()

	const attempts = 6
	for i := 0; i < attempts; i++ {
		body := strings.NewReader(fmt.Sprintf(`{"email": "forgot-%d@example.com"}`, i))
		resp, err := http.Post(testServer.URL+"/api/forgot-password", "application/json", body)
		if err != nil {
			t.Fatalf("Error sending forgot-password request: %v", err)
		}
		resp.Body.Close()

		want := http.StatusOK
		if i == attempts-1 {
			want = http.StatusTooManyRequests
		}
		if resp.StatusCode != want {
			t.Errorf("attempt %d: want status %d, got %d", i+1, want, resp.StatusCode)
		}
	}
}
