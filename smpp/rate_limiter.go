package smpp

import (
	"sync"
	"time"
)

// RateLimiter controls the rate of submissions.
type RateLimiter struct {
	globalLimit    int
	perSessionLimit int
	mu             sync.Mutex
	globalCount    int
	sessionCounts  map[*SingleSession]int
}

// NewRateLimiter creates a new rate limiter.
func NewRateLimiter(globalLimit, perSessionLimit int) *RateLimiter {
	return &RateLimiter{
		globalLimit:    globalLimit,
		perSessionLimit: perSessionLimit,
		sessionCounts:  make(map[*SingleSession]int),
	}
}

// Allow checks if a submission is allowed.
func (r *RateLimiter) Allow(session *SingleSession) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.globalCount >= r.globalLimit {
		return false
	}

	if r.sessionCounts[session] >= r.perSessionLimit {
		return false
	}

	r.globalCount++
	r.sessionCounts[session]++
	return true
}

// Reset resets the counts periodically.
func (r *RateLimiter) Reset() {
	ticker := time.NewTicker(time.Minute)
	go func() {
		for range ticker.C {
			r.mu.Lock()
			r.globalCount = 0
			for k := range r.sessionCounts {
				r.sessionCounts[k] = 0
			}
			r.mu.Unlock()
		}
	}()
}
