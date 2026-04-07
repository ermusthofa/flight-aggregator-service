package ratelimit

import (
	"sync"
	"time"
)

// RateLimiter allows a fixed number of requests within a time window.
type RateLimiter struct {
	mu          sync.Mutex
	lastRequest time.Time
	interval    time.Duration // minimum time between requests
}

// New creates a rate limiter with a given number of requests per duration.
// Examples:
//
//	New(30, time.Minute)  // 30 requests per minute
//	New(10, time.Second)  // 10 requests per second
//	New(100, time.Hour)   // 100 requests per hour
func New(requests int, duration time.Duration) *RateLimiter {
	interval := duration / time.Duration(requests)
	return &RateLimiter{interval: interval}
}

// Allow returns true if the request is allowed.
func (r *RateLimiter) Allow() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	now := time.Now()
	if now.Sub(r.lastRequest) < r.interval {
		return false
	}
	r.lastRequest = now
	return true
}
