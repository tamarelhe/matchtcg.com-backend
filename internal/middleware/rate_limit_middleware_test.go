package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTokenBucket(t *testing.T) {
	t.Run("basic functionality", func(t *testing.T) {
		bucket := NewTokenBucket(5, 2) // 5 capacity, 2 tokens per second

		// Should be able to take 5 tokens initially
		for i := 0; i < 5; i++ {
			assert.True(t, bucket.TakeToken(), "should be able to take token %d", i+1)
		}

		// 6th token should fail
		assert.False(t, bucket.TakeToken(), "should not be able to take 6th token")
	})

	t.Run("refill functionality", func(t *testing.T) {
		bucket := NewTokenBucket(3, 2) // 3 capacity, 2 tokens per second

		// Exhaust all tokens
		for i := 0; i < 3; i++ {
			assert.True(t, bucket.TakeToken())
		}
		assert.False(t, bucket.TakeToken())

		// Wait for refill (1 second = 2 tokens)
		time.Sleep(1100 * time.Millisecond)

		// Should be able to take 2 tokens
		assert.True(t, bucket.TakeToken())
		assert.True(t, bucket.TakeToken())
		assert.False(t, bucket.TakeToken()) // 3rd should fail
	})

	t.Run("capacity limit", func(t *testing.T) {
		bucket := NewTokenBucket(2, 10) // 2 capacity, 10 tokens per second

		// Exhaust tokens
		assert.True(t, bucket.TakeToken())
		assert.True(t, bucket.TakeToken())
		assert.False(t, bucket.TakeToken())

		// Wait for more than enough time to refill
		time.Sleep(1100 * time.Millisecond)

		// Should only have capacity tokens available (2), not 10
		assert.True(t, bucket.TakeToken())
		assert.True(t, bucket.TakeToken())
		assert.False(t, bucket.TakeToken())
	})
}

func TestRateLimitMiddleware_RateLimit(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 2,
		BurstSize:         3,
		CleanupInterval:   time.Minute,
	}

	middleware := NewRateLimitMiddleware(config)
	defer middleware.Close()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := middleware.RateLimit(testHandler)

	t.Run("within rate limit", func(t *testing.T) {
		// Should be able to make 3 requests (burst size)
		for i := 0; i < 3; i++ {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = "192.168.1.1:12345"
			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusOK, rr.Code)
			assert.Equal(t, "success", rr.Body.String())

			// Check rate limit headers
			assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Limit"))
			remaining, err := strconv.Atoi(rr.Header().Get("X-RateLimit-Remaining"))
			require.NoError(t, err)
			assert.Equal(t, 2-i, remaining) // Should decrease with each request
		}
	})

	t.Run("exceeds rate limit", func(t *testing.T) {
		// Make 4th request (should be rate limited)
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.2:12345"
		rr := httptest.NewRecorder()

		// Exhaust burst first
		for i := 0; i < 3; i++ {
			handler.ServeHTTP(rr, req)
			rr = httptest.NewRecorder()
		}

		// This should be rate limited
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
		assert.Contains(t, rr.Body.String(), "Rate limit exceeded")
		assert.Equal(t, "2", rr.Header().Get("X-RateLimit-Limit"))
		assert.Equal(t, "0", rr.Header().Get("X-RateLimit-Remaining"))
		assert.NotEmpty(t, rr.Header().Get("X-RateLimit-Reset"))
	})

	t.Run("different IPs have separate limits", func(t *testing.T) {
		// IP 1 exhausts its limit
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.3:12345"

		for i := 0; i < 3; i++ {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req1)
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		// IP 1 should be rate limited
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)
		assert.Equal(t, http.StatusTooManyRequests, rr1.Code)

		// IP 2 should still work
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.4:12345"
		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)
		assert.Equal(t, http.StatusOK, rr2.Code)
	})
}

func TestRateLimitMiddleware_RateLimitByUser(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 2,
		BurstSize:         2,
		CleanupInterval:   time.Minute,
	}

	middleware := NewRateLimitMiddleware(config)
	defer middleware.Close()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	handler := middleware.RateLimitByUser(testHandler)

	t.Run("authenticated user rate limiting", func(t *testing.T) {
		// Create request with user context
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.5:12345"
		ctx := context.WithValue(req.Context(), UserIDKey, "user123")
		req = req.WithContext(ctx)

		// Should be able to make 2 requests
		for i := 0; i < 2; i++ {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		// 3rd request should be rate limited
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	})

	t.Run("unauthenticated user falls back to IP", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.6:12345"

		// Should be able to make 2 requests
		for i := 0; i < 2; i++ {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		// 3rd request should be rate limited
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusTooManyRequests, rr.Code)
	})

	t.Run("different users have separate limits", func(t *testing.T) {
		// User 1 exhausts limit
		req1 := httptest.NewRequest("GET", "/test", nil)
		req1.RemoteAddr = "192.168.1.7:12345"
		ctx1 := context.WithValue(req1.Context(), UserIDKey, "user1")
		req1 = req1.WithContext(ctx1)

		for i := 0; i < 2; i++ {
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req1)
			assert.Equal(t, http.StatusOK, rr.Code)
		}

		// User 1 should be rate limited
		rr1 := httptest.NewRecorder()
		handler.ServeHTTP(rr1, req1)
		assert.Equal(t, http.StatusTooManyRequests, rr1.Code)

		// User 2 should still work
		req2 := httptest.NewRequest("GET", "/test", nil)
		req2.RemoteAddr = "192.168.1.8:12345"
		ctx2 := context.WithValue(req2.Context(), UserIDKey, "user2")
		req2 = req2.WithContext(ctx2)

		rr2 := httptest.NewRecorder()
		handler.ServeHTTP(rr2, req2)
		assert.Equal(t, http.StatusOK, rr2.Code)
	})
}

/* TODO
func TestRateLimitMiddleware_Cleanup(t *testing.T) {
	config := RateLimitConfig{
		RequestsPerSecond: 10,
		BurstSize:         10,
		CleanupInterval:   50 * time.Millisecond,
	}

	middleware := NewRateLimitMiddleware(config)
	defer middleware.Close()

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.RateLimit(testHandler)

	// Make requests from different IPs to create buckets
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1." + strconv.Itoa(i+10) + ":12345"
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
	}

	// Check that buckets were created
	middleware.mu.RLock()
	initialCount := len(middleware.buckets)
	middleware.mu.RUnlock()
	assert.Equal(t, 5, initialCount)

	// Wait for cleanup to run multiple times
	time.Sleep(200 * time.Millisecond)

	// Buckets should still exist (they're not old enough)
	middleware.mu.RLock()
	afterCleanupCount := len(middleware.buckets)
	middleware.mu.RUnlock()
	assert.Equal(t, 5, afterCleanupCount)
}
*/

func TestGetClientIP(t *testing.T) {
	t.Run("X-Forwarded-For header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
		req.RemoteAddr = "192.168.1.1:12345"

		ip := getClientIP(req)
		assert.Equal(t, "203.0.113.1", ip)
	})

	t.Run("X-Real-IP header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Real-IP", "203.0.113.2")
		req.RemoteAddr = "192.168.1.1:12345"

		ip := getClientIP(req)
		assert.Equal(t, "203.0.113.2", ip)
	})

	t.Run("RemoteAddr fallback", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"

		ip := getClientIP(req)
		assert.Equal(t, "192.168.1.1", ip)
	})

	t.Run("RemoteAddr without port", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1"

		ip := getClientIP(req)
		assert.Equal(t, "192.168.1.1", ip)
	})
}
