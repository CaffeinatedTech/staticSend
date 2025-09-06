package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(time.Second, 5)

	if limiter.rate != time.Second {
		t.Errorf("Expected rate %v, got %v", time.Second, limiter.rate)
	}

	if limiter.burst != 5 {
		t.Errorf("Expected burst %d, got %d", 5, limiter.burst)
	}

	if limiter.buckets == nil {
		t.Error("Expected buckets map to be initialized")
	}
}

func TestRateLimiter_Limit(t *testing.T) {
	limiter := NewRateLimiter(100*time.Millisecond, 2) // 10 requests per second, burst of 2

	// First two requests should not be limited
	if limiter.Limit("test-key") {
		t.Error("First request should not be limited")
	}

	if limiter.Limit("test-key") {
		t.Error("Second request should not be limited")
	}

	// Third request should be limited (burst exceeded)
	if !limiter.Limit("test-key") {
		t.Error("Third request should be limited")
	}

	// Wait for tokens to replenish
	time.Sleep(150 * time.Millisecond)

	// Should be able to make one more request
	if limiter.Limit("test-key") {
		t.Error("Request after wait should not be limited")
	}
}

func TestRateLimiter_DifferentKeys(t *testing.T) {
	limiter := NewRateLimiter(time.Second, 1)

	// Different keys should have separate buckets
	if limiter.Limit("key1") {
		t.Error("First request for key1 should not be limited")
	}

	if limiter.Limit("key2") {
		t.Error("First request for key2 should not be limited")
	}

	// Second requests should be limited for both keys
	if !limiter.Limit("key1") {
		t.Error("Second request for key1 should be limited")
	}

	if !limiter.Limit("key2") {
		t.Error("Second request for key2 should be limited")
	}
}

func TestIPRateLimit_Middleware(t *testing.T) {
	// Create test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply rate limiting middleware (1 request per second, burst of 1)
	middleware := IPRateLimit(time.Second, 1)(handler)

	// First request should succeed
	req := httptest.NewRequest("GET", "/", nil)
	req.RemoteAddr = "192.168.1.1:8080"

	rr := httptest.NewRecorder()
	middleware.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rr.Code)
	}

	// Second request should be rate limited
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req)

	if rr2.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", rr2.Code)
	}
}

func TestGetClientIP(t *testing.T) {
	tests := []struct {
		name     string
		header   map[string]string
		expected string
	}{
		{
			name:     "CF-Connecting-IP header",
			header:   map[string]string{"CF-Connecting-IP": "1.2.3.4"},
			expected: "1.2.3.4",
		},
		{
			name:     "X-Forwarded-For header",
			header:   map[string]string{"X-Forwarded-For": "5.6.7.8"},
			expected: "5.6.7.8",
		},
		{
			name:     "X-Real-IP header",
			header:   map[string]string{"X-Real-IP": "9.10.11.12"},
			expected: "9.10.11.12",
		},
		{
			name:     "Remote address fallback",
			header:   map[string]string{},
			expected: "192.168.1.1:8080",
		},
		{
			name: "Precedence - CF-Connecting-IP first",
			header: map[string]string{
				"CF-Connecting-IP": "1.2.3.4",
				"X-Forwarded-For":  "5.6.7.8",
				"X-Real-IP":        "9.10.11.12",
			},
			expected: "1.2.3.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			req.RemoteAddr = "192.168.1.1:8080"

			// Set headers
			for key, value := range tt.header {
				req.Header.Set(key, value)
			}

			ip := getClientIP(req)
			if ip != tt.expected {
				t.Errorf("Expected IP %s, got %s", tt.expected, ip)
			}
		})
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	limiter := NewRateLimiter(time.Second, 1)

	// Add a bucket
	limiter.Limit("test-key")

	// Verify bucket exists
	limiter.mu.Lock()
	if _, exists := limiter.buckets["test-key"]; !exists {
		t.Error("Bucket should exist")
	}
	limiter.mu.Unlock()

	// Manually set last check to be old
	limiter.mu.Lock()
	limiter.buckets["test-key"].LastCheck = time.Now().Add(-2 * time.Hour)
	limiter.mu.Unlock()

	// Trigger cleanup
	limiter.mu.Lock()
	limiter.cleanup()
	limiter.mu.Unlock()

	// Bucket should be removed
	limiter.mu.Lock()
	_, exists := limiter.buckets["test-key"]
	limiter.mu.Unlock()

	if exists {
		t.Error("Bucket should have been cleaned up")
	}
}

func TestRateLimiter_ConcurrentAccess(t *testing.T) {
	limiter := NewRateLimiter(time.Millisecond, 100)

	// Test concurrent access from multiple goroutines
	const goroutines = 10
	const requests = 20

	done := make(chan bool)
	limitedCount := 0

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < requests; j++ {
				key := "test-key"
				if limiter.Limit(key) {
					limitedCount++
				}
				time.Sleep(time.Microsecond * 10)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < goroutines; i++ {
		<-done
	}

	// Should have some limited requests due to rate limiting
	if limitedCount == 0 {
		t.Error("Expected some requests to be limited under concurrent access")
	}
}
