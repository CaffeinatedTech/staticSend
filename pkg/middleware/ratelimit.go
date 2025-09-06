package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

// RateLimiter implements a token bucket rate limiter
type RateLimiter struct {
	mu          sync.Mutex
	rate        time.Duration
	burst       int
	buckets     map[string]*tokenBucket
	cleanupTime time.Time
}

// tokenBucket represents a token bucket for a specific key (e.g., IP address)
type tokenBucket struct {
	Tokens    int
	LastCheck time.Time
}

// NewRateLimiter creates a new rate limiter with the specified rate and burst capacity
func NewRateLimiter(rate time.Duration, burst int) *RateLimiter {
	return &RateLimiter{
		rate:    rate,
		burst:   burst,
		buckets: make(map[string]*tokenBucket),
	}
}

// Limit returns true if the request should be rate limited
func (rl *RateLimiter) Limit(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Clean up old buckets periodically
	now := time.Now()
	if now.After(rl.cleanupTime) {
		rl.cleanup()
		// Clean up every 5 minutes
		rl.cleanupTime = now.Add(5 * time.Minute)
	}

	bucket, exists := rl.buckets[key]
	if !exists {
		bucket = &tokenBucket{
			Tokens:    rl.burst,
			LastCheck: now,
		}
		rl.buckets[key] = bucket
	}

	// Calculate how many tokens to add based on elapsed time
	elapsed := now.Sub(bucket.LastCheck)
	tokensToAdd := int(elapsed / rl.rate)

	if tokensToAdd > 0 {
		bucket.Tokens += tokensToAdd
		if bucket.Tokens > rl.burst {
			bucket.Tokens = rl.burst
		}
		bucket.LastCheck = now
	}

	// Check if we have tokens available
	if bucket.Tokens <= 0 {
		return true
	}

	// Consume a token
	bucket.Tokens--
	return false
}

// cleanup removes old buckets to prevent memory leaks
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	for key, bucket := range rl.buckets {
		// Remove buckets that haven't been used in the last hour
		if now.Sub(bucket.LastCheck) > time.Hour {
			delete(rl.buckets, key)
		}
	}
}

// IPRateLimit creates a middleware that rate limits by IP address
func IPRateLimit(rate time.Duration, burst int) func(http.Handler) http.Handler {
	limiter := NewRateLimiter(rate, burst)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get client IP
			ip := getClientIP(r)

			// Check rate limit
			if limiter.Limit(ip) {
				http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP address from the request
func getClientIP(r *http.Request) string {
	// Check Cloudflare headers first
	if ip := r.Header.Get("CF-Connecting-IP"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Forwarded-For"); ip != "" {
		return ip
	}
	if ip := r.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to remote address
	return r.RemoteAddr
}

// RateLimitResponse adds rate limit headers to responses
func RateLimitResponse(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a custom response writer to capture status code
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			// Add rate limit headers for successful requests
			if ww.Status() >= 200 && ww.Status() < 300 {
				ip := getClientIP(r)

				limiter.mu.Lock()
				bucket, exists := limiter.buckets[ip]
				limiter.mu.Unlock()

				if exists {
					// Calculate remaining tokens and reset time
					remaining := bucket.Tokens
					resetTime := bucket.LastCheck.Add(time.Duration(limiter.burst-bucket.Tokens) * limiter.rate)

					w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", limiter.burst))
					w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
					w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime.Unix()))
				}
			}
		})
	}
}
