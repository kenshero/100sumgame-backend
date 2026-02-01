package service

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// OperationRateLimiter implements rate limiting for specific operations
type OperationRateLimiter struct {
	guestRequests map[string][]time.Time
	ipRequests    map[string][]time.Time
	mu            sync.RWMutex
	rate          int
	window        time.Duration
}

// NewOperationRateLimiter creates a new operation rate limiter
func NewOperationRateLimiter(rate int, window time.Duration) *OperationRateLimiter {
	rl := &OperationRateLimiter{
		guestRequests: make(map[string][]time.Time),
		ipRequests:    make(map[string][]time.Time),
		rate:          rate,
		window:        window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// AllowGuest checks if a guest is allowed to perform an operation
func (rl *OperationRateLimiter) AllowGuest(guestID uuid.UUID) bool {
	return rl.Allow(guestID.String(), "guest")
}

// AllowIP checks if an IP is allowed to perform an operation
func (rl *OperationRateLimiter) AllowIP(ip string) bool {
	return rl.Allow(ip, "ip")
}

// Allow checks if an identifier is allowed to perform an operation
func (rl *OperationRateLimiter) Allow(identifier string, identifierType string) bool {
	var requests []time.Time

	rl.mu.RLock()
	if identifierType == "guest" {
		requests = rl.guestRequests[identifier]
	} else {
		requests = rl.ipRequests[identifier]
	}
	rl.mu.RUnlock()

	now := time.Now()
	cutoff := now.Add(-rl.window)

	// Filter old requests
	validRequests := make([]time.Time, 0)
	for _, reqTime := range requests {
		if reqTime.After(cutoff) {
			validRequests = append(validRequests, reqTime)
		}
	}

	// Check if rate limit exceeded
	if len(validRequests) >= rl.rate {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)

	// Update store
	rl.mu.Lock()
	if identifierType == "guest" {
		rl.guestRequests[identifier] = validRequests
	} else {
		rl.ipRequests[identifier] = validRequests
	}
	rl.mu.Unlock()

	return true
}

// cleanup removes old entries to prevent memory leaks
func (rl *OperationRateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		cutoff := now.Add(-10 * time.Minute)

		// Clean guest requests
		for id, requests := range rl.guestRequests {
			validRequests := make([]time.Time, 0)
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}
			if len(validRequests) == 0 {
				delete(rl.guestRequests, id)
			} else {
				rl.guestRequests[id] = validRequests
			}
		}

		// Clean IP requests
		for id, requests := range rl.ipRequests {
			validRequests := make([]time.Time, 0)
			for _, reqTime := range requests {
				if reqTime.After(cutoff) {
					validRequests = append(validRequests, reqTime)
				}
			}
			if len(validRequests) == 0 {
				delete(rl.ipRequests, id)
			} else {
				rl.ipRequests[id] = validRequests
			}
		}

		rl.mu.Unlock()
	}
}
