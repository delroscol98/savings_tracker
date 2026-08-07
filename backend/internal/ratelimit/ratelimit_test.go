package ratelimit

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) current() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

func TestRateLimiterAllow(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	rl := New(2, time.Minute)
	rl.now = clock.current

	if !rl.Allow("1.1.1.1") {
		t.Error("first request within limit should be allowed")
	}
	if !rl.Allow("1.1.1.1") {
		t.Error("second request at the limit should be allowed")
	}
	if rl.Allow("1.1.1.1") {
		t.Error("third request over the limit should be denied")
	}
	if !rl.Allow("2.2.2.2") {
		t.Error("requests from a distinct key should have their own budget")
	}
}

func TestRateLimiterWindowReset(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	rl := New(2, 10*time.Minute)
	rl.now = clock.current

	rl.Allow("1.1.1.1")
	rl.Allow("1.1.1.1")
	if rl.Allow("1.1.1.1") {
		t.Fatal("request over the limit should be denied before window reset")
	}

	clock.advance(11 * time.Minute)

	if !rl.Allow("1.1.1.1") {
		t.Error("request after window expiry should be allowed again")
	}
}

func TestRateLimiterReset(t *testing.T) {
	rl := New(1, time.Minute)
	rl.Allow("1.1.1.1")
	if rl.Allow("1.1.1.1") {
		t.Fatal("request over the limit should be denied")
	}

	rl.Reset()

	if !rl.Allow("1.1.1.1") {
		t.Error("request after Reset should be allowed again")
	}
}

func TestRateLimiterEviction(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	rl := New(1, 10*time.Minute)
	rl.now = clock.current
	rl.maxEntries = 1

	rl.Allow("1.1.1.1")
	rl.Allow("2.2.2.2")
	if got := len(rl.entries); got != 2 {
		t.Fatalf("unexpired entries should remain, got %d", got)
	}

	clock.advance(11 * time.Minute)
	rl.Allow("3.3.3.3")

	if got := len(rl.entries); got != 1 {
		t.Fatalf("expired entries should be swept, got %d", got)
	}
	if _, ok := rl.entries["3.3.3.3"]; !ok {
		t.Error("newest entry should be retained after sweep")
	}
}

func TestRateLimiterConcurrent(t *testing.T) {
	rl := New(1, time.Minute)

	const goroutines = 100
	var wg sync.WaitGroup
	allowed := make(chan bool, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed <- rl.Allow("1.1.1.1")
		}()
	}
	wg.Wait()
	close(allowed)

	successes := 0
	for a := range allowed {
		if a {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("exactly one concurrent request should be allowed, got %d", successes)
	}
}

func TestLoginThrottlerLockout(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	lt := NewLoginThrottler(5, 10*time.Minute)
	lt.now = clock.current

	email := "test@example.com"

	if lt.IsLockedOut(email) {
		t.Fatal("account should not be locked out before any failures")
	}

	for i := 0; i < 4; i++ {
		lt.RecordFailure(email)
		if lt.IsLockedOut(email) {
			t.Fatalf("account should not be locked out after %d failures", i+1)
		}
	}

	lt.RecordFailure(email)
	if !lt.IsLockedOut(email) {
		t.Fatal("account should be locked out after reaching max failures")
	}

	lt.Clear(email)
	if lt.IsLockedOut(email) {
		t.Fatal("account should not be locked out after Clear")
	}
}

func TestLoginThrottlerWindowReset(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	lt := NewLoginThrottler(5, 10*time.Minute)
	lt.now = clock.current

	email := "test@example.com"
	for i := 0; i < 5; i++ {
		lt.RecordFailure(email)
	}
	if !lt.IsLockedOut(email) {
		t.Fatal("account should be locked out")
	}

	clock.advance(11 * time.Minute)

	if lt.IsLockedOut(email) {
		t.Fatal("account should be unlocked after window expiry")
	}

	lt.RecordFailure(email)
	if lt.IsLockedOut(email) {
		t.Fatal("failure count should have reset after window expiry")
	}
}

func TestLoginThrottlerEviction(t *testing.T) {
	clock := &fakeClock{now: time.Unix(0, 0)}
	lt := NewLoginThrottler(1, 10*time.Minute)
	lt.now = clock.current
	lt.maxEntries = 1

	lt.RecordFailure("one@example.com")
	lt.RecordFailure("two@example.com")
	if got := len(lt.entries); got != 2 {
		t.Fatalf("unexpired entries should remain, got %d", got)
	}

	clock.advance(11 * time.Minute)
	lt.RecordFailure("three@example.com")

	if got := len(lt.entries); got != 1 {
		t.Fatalf("expired entries should be swept, got %d", got)
	}
	if _, ok := lt.entries["three@example.com"]; !ok {
		t.Error("newest entry should be retained after sweep")
	}
}

func TestLoginThrottlerConcurrent(t *testing.T) {
	lt := NewLoginThrottler(5, time.Minute)

	email := "test@example.com"
	const goroutines = 100
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lt.RecordFailure(email)
		}()
	}
	wg.Wait()

	lt.mu.Lock()
	count := lt.entries[email].failures
	lt.mu.Unlock()

	if count != goroutines {
		t.Fatalf("all concurrent failures should be recorded, got %d", count)
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		xff        string
		want       string
	}{
		{
			name:       "no x-forwarded-for",
			remoteAddr: "10.0.0.1:8080",
			want:       "10.0.0.1",
		},
		{
			name:       "leftmost x-forwarded-for",
			remoteAddr: "10.0.0.1:8080",
			xff:        "1.2.3.4, 5.6.7.8",
			want:       "1.2.3.4",
		},
		{
			name:       "skips invalid x-forwarded-for",
			remoteAddr: "10.0.0.1:8080",
			xff:        "not-an-ip, 9.9.9.9",
			want:       "9.9.9.9",
		},
		{
			name:       "invalid x-forwarded-for falls back to remote addr",
			remoteAddr: "10.0.0.1:8080",
			xff:        "garbage",
			want:       "10.0.0.1",
		},
		{
			name:       "invalid remote addr",
			remoteAddr: "garbage",
			want:       "unknown",
		},
		{
			name:       "ipv6 remote addr",
			remoteAddr: "[::1]:8080",
			want:       "::1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}

			if got := ClientIP(r); got != tt.want {
				t.Errorf("want %q, got %q", tt.want, got)
			}
		})
	}
}

func TestClientIPNilRequest(t *testing.T) {
	if got := ClientIP(nil); got != "unknown" {
		t.Errorf("want %q, got %q", "unknown", got)
	}
}
