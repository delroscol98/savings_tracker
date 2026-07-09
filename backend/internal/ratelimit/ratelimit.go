package ratelimit

import (
	"sync"
	"time"
)

type entry struct {
	count    int
	windowAt time.Time
}

type RateLimiter struct {
	mu      sync.Mutex
	entries map[string]*entry
	limit   int
	window  time.Duration
}

func New(limit int, window time.Duration) *RateLimiter {
	rateLimiter := RateLimiter{
		entries: map[string]*entry{},
		limit:   limit,
		window:  window,
	}

	return &rateLimiter
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	e, ok := rl.entries[ip]
	if !ok {
		e := &entry{
			count:    1,
			windowAt: time.Now(),
		}
		rl.entries[ip] = e

		return e.count <= rl.limit
	}

	if time.Now().After(e.windowAt.Add(rl.window)) {
		e.count = 0
		e.windowAt = time.Now()
	}
	e.count++

	return e.count <= rl.limit
}
