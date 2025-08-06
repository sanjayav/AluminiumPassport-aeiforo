package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"aluminium-passport/internal/config"
)

// RateLimiter holds rate limiting data
type RateLimiter struct {
	requests map[string][]time.Time
	mutex    sync.RWMutex
	limit    int
	window   time.Duration
}

var globalRateLimiter *RateLimiter
var once sync.Once

// initRateLimiter initializes the global rate limiter
func initRateLimiter() {
	cfg := config.AppConfig
	globalRateLimiter = &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    cfg.RateLimitRPM,
		window:   time.Minute,
	}

	// Start cleanup goroutine
	go globalRateLimiter.cleanup()
}

// RateLimitMiddleware applies rate limiting to requests
func RateLimitMiddleware(next http.Handler) http.Handler {
	once.Do(initRateLimiter)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client identifier (IP address)
		clientIP := getClientIP(r)

		// Check if request is allowed
		if !globalRateLimiter.isAllowed(clientIP) {
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", globalRateLimiter.limit))
			w.Header().Set("X-RateLimit-Window", globalRateLimiter.window.String())
			w.Header().Set("Retry-After", "60")

			http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers
		remaining := globalRateLimiter.getRemainingRequests(clientIP)
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", globalRateLimiter.limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		w.Header().Set("X-RateLimit-Window", globalRateLimiter.window.String())

		next.ServeHTTP(w, r)
	})
}

// isAllowed checks if a request from the given client is allowed
func (rl *RateLimiter) isAllowed(clientID string) bool {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this client
	requests := rl.requests[clientID]

	// Filter out requests outside the current window
	validRequests := make([]time.Time, 0, len(requests))
	for _, req := range requests {
		if req.After(windowStart) {
			validRequests = append(validRequests, req)
		}
	}

	// Check if limit is exceeded
	if len(validRequests) >= rl.limit {
		return false
	}

	// Add current request
	validRequests = append(validRequests, now)
	rl.requests[clientID] = validRequests

	return true
}

// getRemainingRequests returns the number of remaining requests for a client
func (rl *RateLimiter) getRemainingRequests(clientID string) int {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	requests := rl.requests[clientID]
	validCount := 0

	for _, req := range requests {
		if req.After(windowStart) {
			validCount++
		}
	}

	remaining := rl.limit - validCount
	if remaining < 0 {
		remaining = 0
	}

	return remaining
}

// cleanup removes old entries from the rate limiter
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		rl.mutex.Lock()

		now := time.Now()
		windowStart := now.Add(-rl.window)

		for clientID, requests := range rl.requests {
			validRequests := make([]time.Time, 0, len(requests))

			for _, req := range requests {
				if req.After(windowStart) {
					validRequests = append(validRequests, req)
				}
			}

			if len(validRequests) == 0 {
				delete(rl.requests, clientID)
			} else {
				rl.requests[clientID] = validRequests
			}
		}

		rl.mutex.Unlock()
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP if multiple are present
		if idx := len(xff); idx > 0 {
			if commaIdx := 0; commaIdx < idx {
				for i, char := range xff {
					if char == ',' {
						commaIdx = i
						break
					}
				}
				if commaIdx > 0 {
					return xff[:commaIdx]
				}
			}
			return xff
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	return r.RemoteAddr
}

// NewRateLimiter creates a new rate limiter with custom settings
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	go rl.cleanup()
	return rl
}

// CustomRateLimitMiddleware creates a rate limiting middleware with custom settings
func CustomRateLimitMiddleware(limit int, window time.Duration) func(http.Handler) http.Handler {
	rateLimiter := NewRateLimiter(limit, window)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			clientIP := getClientIP(r)

			if !rateLimiter.isAllowed(clientIP) {
				w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
				w.Header().Set("X-RateLimit-Window", window.String())
				w.Header().Set("Retry-After", fmt.Sprintf("%.0f", window.Seconds()))

				http.Error(w, "Rate limit exceeded. Please try again later.", http.StatusTooManyRequests)
				return
			}

			remaining := rateLimiter.getRemainingRequests(clientIP)
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limit))
			w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
			w.Header().Set("X-RateLimit-Window", window.String())

			next.ServeHTTP(w, r)
		})
	}
}

// PerUserRateLimitMiddleware applies different rate limits based on user role
func PerUserRateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get user from context (if authenticated)
		var limit int
		var window time.Duration

		if claims, ok := GetUserFromContext(r); ok {
			// Set different limits based on user role
			switch claims.Role {
			case "admin":
				limit = 1000 // High limit for admins
				window = time.Minute
			case "certifier", "auditor":
				limit = 500 // Medium limit for certifiers/auditors
				window = time.Minute
			case "miner", "manufacturer", "recycler":
				limit = 200 // Standard limit for operational roles
				window = time.Minute
			default:
				limit = 100 // Basic limit for viewers
				window = time.Minute
			}
		} else {
			// Unauthenticated users get lowest limit
			limit = 50
			window = time.Minute
		}

		// Apply custom rate limiting
		CustomRateLimitMiddleware(limit, window)(next).ServeHTTP(w, r)
	})
}
