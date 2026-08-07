package ratelimit

import (
	"net/http"
	"net/netip"
	"strings"
	"sync"
	"time"
)

const defaultMaxEntries = 1024

type entry struct {
	count    int
	windowAt time.Time
}

type RateLimiter struct {
	mu         sync.Mutex
	entries    map[string]*entry
	limit      int
	window     time.Duration
	maxEntries int
	now        func() time.Time
}

func New(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		entries:    map[string]*entry{},
		limit:      limit,
		window:     window,
		maxEntries: defaultMaxEntries,
		now:        time.Now,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	e, ok := rl.entries[ip]
	if !ok {
		e = &entry{
			count:    1,
			windowAt: rl.now(),
		}
		rl.entries[ip] = e
	} else {
		if rl.now().After(e.windowAt.Add(rl.window)) {
			e.count = 0
			e.windowAt = rl.now()
		}
		e.count++
	}

	rl.sweep()

	return e.count <= rl.limit
}

func (rl *RateLimiter) sweep() {
	if len(rl.entries) <= rl.maxEntries {
		return
	}

	now := rl.now()
	for key, e := range rl.entries {
		if now.After(e.windowAt.Add(rl.window)) {
			delete(rl.entries, key)
		}
	}
}

func (rl *RateLimiter) Reset() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	rl.entries = make(map[string]*entry)
}

// ClientIP returns the client's IP address. When present it trusts the
// leftmost X-Forwarded-For value (set by the trusted proxy) and falls back
// to RemoteAddr otherwise.
func ClientIP(r *http.Request) string {
	if r == nil {
		return "unknown"
	}

	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		for _, part := range strings.Split(xff, ",") {
			ipStr := strings.TrimSpace(part)
			if ip, err := netip.ParseAddr(ipStr); err == nil {
				return ip.String()
			}
		}
	}

	ip := "unknown"
	ipPort, err := netip.ParseAddrPort(r.RemoteAddr)
	if err == nil {
		ip = ipPort.Addr().String()
	}

	return ip
}
