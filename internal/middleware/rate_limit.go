package middleware

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int // requests per minute
	window   time.Duration
}

// Visitor tracks requests from a specific IP
type Visitor struct {
	requests []time.Time
	mu       sync.Mutex
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.RLock()
	visitor, exists := rl.visitors[ip]
	rl.mu.RUnlock()

	if !exists {
		rl.mu.Lock()
		// Double-check after acquiring write lock
		visitor, exists = rl.visitors[ip]
		if !exists {
			visitor = &Visitor{
				requests: make([]time.Time, 0),
			}
			rl.visitors[ip] = visitor
		}
		rl.mu.Unlock()
	}

	visitor.mu.Lock()
	defer visitor.mu.Unlock()

	now := time.Now()
	// Remove old requests outside the time window
	cutoff := now.Add(-rl.window)
	validRequests := make([]time.Time, 0, len(visitor.requests))
	for _, reqTime := range visitor.requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if rate limit exceeded
	if len(validRequests) >= rl.rate {
		return false
	}

	// Add current request
	visitor.requests = append(validRequests, now)
	return true
}

// cleanup removes old visitor entries to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		for ip, visitor := range rl.visitors {
			visitor.mu.Lock()
			cutoff := now.Add(-5 * time.Minute) // Keep entries for 5 minutes of inactivity
			hasRecent := false
			for _, reqTime := range visitor.requests {
				if reqTime.After(cutoff) {
					hasRecent = true
					break
				}
			}
			visitor.mu.Unlock()

			if !hasRecent {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a middleware for rate limiting
func RateLimitMiddleware(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get real IP from X-Real-IP or X-Forwarded-For header if behind proxy
			ip := r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				ip = r.RemoteAddr
			}

			if !limiter.Allow(ip) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusTooManyRequests)
				w.Write([]byte(`{"error":"Rate limit exceeded. Please try again later."}`))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
