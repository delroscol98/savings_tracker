package ratelimit

import (
	"sync"
	"time"
)

type throttleEntry struct {
	failures int
	windowAt time.Time
}

type LoginThrottler struct {
	mu          sync.Mutex
	entries     map[string]*throttleEntry
	maxFailures int
	window      time.Duration
	maxEntries  int
	now         func() time.Time
}

func NewLoginThrottler(maxFailures int, window time.Duration) *LoginThrottler {
	return &LoginThrottler{
		entries:     map[string]*throttleEntry{},
		maxFailures: maxFailures,
		window:      window,
		maxEntries:  defaultMaxEntries,
		now:         time.Now,
	}
}

// IsLockedOut reports whether the account has accumulated maxFailures or
// more failed logins within the window. Expired entries are cleared.
func (lt *LoginThrottler) IsLockedOut(email string) bool {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	e, ok := lt.entries[email]
	if !ok {
		return false
	}

	if lt.now().After(e.windowAt.Add(lt.window)) {
		delete(lt.entries, email)
		return false
	}

	return e.failures >= lt.maxFailures
}

func (lt *LoginThrottler) RecordFailure(email string) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	e, ok := lt.entries[email]
	if !ok {
		e = &throttleEntry{
			failures: 1,
			windowAt: lt.now(),
		}
		lt.entries[email] = e
	} else {
		if lt.now().After(e.windowAt.Add(lt.window)) {
			e.failures = 0
			e.windowAt = lt.now()
		}
		e.failures++
	}

	lt.sweep()
}

func (lt *LoginThrottler) Clear(email string) {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	delete(lt.entries, email)
}

func (lt *LoginThrottler) Reset() {
	lt.mu.Lock()
	defer lt.mu.Unlock()

	lt.entries = make(map[string]*throttleEntry)
}

func (lt *LoginThrottler) sweep() {
	if len(lt.entries) <= lt.maxEntries {
		return
	}

	now := lt.now()
	for key, e := range lt.entries {
		if now.After(e.windowAt.Add(lt.window)) {
			delete(lt.entries, key)
		}
	}
}
