package middleware

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// TokenBucket represents a token bucket for rate limiting
type TokenBucket struct {
	capacity   int        // Maximum number of tokens
	tokens     int        // Current number of tokens
	refillRate int        // Tokens added per second
	lastRefill time.Time  // Last time bucket was refilled
	mu         sync.Mutex // Mutex for thread safety
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// TakeToken attempts to take a token from the bucket
func (tb *TokenBucket) TakeToken() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * tb.refillRate

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Check if token is available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// RateLimitConfig holds configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerSecond int           // Number of requests allowed per second
	BurstSize         int           // Maximum burst size
	CleanupInterval   time.Duration // How often to clean up old buckets
}

// RateLimitMiddleware handles rate limiting using token bucket algorithm
type RateLimitMiddleware struct {
	config  RateLimitConfig
	buckets map[string]*TokenBucket
	mu      sync.RWMutex
	stopCh  chan struct{}
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(config RateLimitConfig) *RateLimitMiddleware {
	if config.CleanupInterval == 0 {
		config.CleanupInterval = 5 * time.Minute
	}

	middleware := &RateLimitMiddleware{
		config:  config,
		buckets: make(map[string]*TokenBucket),
		stopCh:  make(chan struct{}),
	}

	// Start cleanup goroutine
	go middleware.cleanupBuckets()

	return middleware
}

// RateLimit middleware that applies rate limiting per IP address
func (m *RateLimitMiddleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get client IP
		clientIP := getClientIP(r)

		// Get or create bucket for this IP
		bucket := m.getBucket(clientIP)

		// Try to take a token
		if !bucket.TakeToken() {
			// Rate limit exceeded
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.RequestsPerSecond))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.RequestsPerSecond))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", bucket.tokens))

		// Continue with request
		next.ServeHTTP(w, r)
	})
}

// RateLimitByUser middleware that applies rate limiting per authenticated user
func (m *RateLimitMiddleware) RateLimitByUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to get user ID from context (set by auth middleware)
		userID, hasUser := GetUserID(r)

		// Fall back to IP if no user
		key := getClientIP(r)
		if hasUser {
			key = "user:" + userID
		}

		// Get or create bucket for this key
		bucket := m.getBucket(key)

		// Try to take a token
		if !bucket.TakeToken() {
			// Rate limit exceeded
			w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.RequestsPerSecond))
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Second).Unix()))
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		// Add rate limit headers
		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", m.config.RequestsPerSecond))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", bucket.tokens))

		// Continue with request
		next.ServeHTTP(w, r)
	})
}

// getBucket gets or creates a token bucket for the given key
func (m *RateLimitMiddleware) getBucket(key string) *TokenBucket {
	m.mu.RLock()
	bucket, exists := m.buckets[key]
	m.mu.RUnlock()

	if exists {
		return bucket
	}

	// Create new bucket
	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if bucket, exists := m.buckets[key]; exists {
		return bucket
	}

	bucket = NewTokenBucket(m.config.BurstSize, m.config.RequestsPerSecond)
	m.buckets[key] = bucket
	return bucket
}

// cleanupBuckets removes old unused buckets
func (m *RateLimitMiddleware) cleanupBuckets() {
	ticker := time.NewTicker(m.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.performCleanup()
		case <-m.stopCh:
			return
		}
	}
}

// performCleanup removes buckets that haven't been used recently
func (m *RateLimitMiddleware) performCleanup() {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-2 * m.config.CleanupInterval)

	for key, bucket := range m.buckets {
		bucket.mu.Lock()
		if bucket.lastRefill.Before(cutoff) {
			delete(m.buckets, key)
		}
		bucket.mu.Unlock()
	}
}

// Close stops the cleanup goroutine
func (m *RateLimitMiddleware) Close() {
	close(m.stopCh)
}

// getClientIP extracts the real client IP from the request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header (from load balancers/proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ip := strings.Split(xff, ",")[0]; ip != "" {
			return strings.TrimSpace(ip)
		}
	}

	// Check X-Real-IP header (from nginx)
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fall back to remote address
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
