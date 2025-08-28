package middleware

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/oklog/ulid/v2"
)

// LoggingConfig holds configuration for logging middleware
type LoggingConfig struct {
	Logger       *log.Logger
	SkipPaths    []string // Paths to skip logging (e.g., health checks)
	LogUserAgent bool     // Whether to log user agent
	LogReferer   bool     // Whether to log referer
	LogRequestID bool     // Whether to log request ID
}

// LoggingMiddleware handles request logging
type LoggingMiddleware struct {
	config LoggingConfig
}

// NewLoggingMiddleware creates a new logging middleware
func NewLoggingMiddleware(config LoggingConfig) *LoggingMiddleware {
	if config.Logger == nil {
		config.Logger = log.Default()
	}
	return &LoggingMiddleware{
		config: config,
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and size
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func (rw *responseWriter) WriteHeader(statusCode int) {
	rw.statusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.size += size
	return size, err
}

// Logging middleware that logs HTTP requests
func (m *LoggingMiddleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if path should be skipped
		for _, skipPath := range m.config.SkipPaths {
			if r.URL.Path == skipPath {
				next.ServeHTTP(w, r)
				return
			}
		}

		start := time.Now()

		// Wrap response writer to capture status and size
		wrapped := &responseWriter{
			ResponseWriter: w,
			statusCode:     http.StatusOK, // Default status
		}

		// Process request
		next.ServeHTTP(wrapped, r)

		// Calculate duration
		duration := time.Since(start)

		// Build log message
		logMsg := buildLogMessage(r, wrapped, duration, m.config)

		// Log the request
		m.config.Logger.Println(logMsg)
	})
}

// buildLogMessage constructs the log message
func buildLogMessage(r *http.Request, rw *responseWriter, duration time.Duration, config LoggingConfig) string {
	// Basic format: METHOD PATH STATUS SIZE DURATION IP
	msg := fmt.Sprintf("%s %s %d %d %v %s",
		r.Method,
		r.URL.Path,
		rw.statusCode,
		rw.size,
		duration,
		getClientIP(r),
	)

	// Add query parameters if present
	if r.URL.RawQuery != "" {
		msg += " query=" + r.URL.RawQuery
	}

	// Add user agent if configured
	if config.LogUserAgent {
		if ua := r.Header.Get("User-Agent"); ua != "" {
			msg += " ua=\"" + ua + "\""
		}
	}

	// Add referer if configured
	if config.LogReferer {
		if referer := r.Header.Get("Referer"); referer != "" {
			msg += " referer=\"" + referer + "\""
		}
	}

	// Add request ID if configured and present
	if config.LogRequestID {
		if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
			msg += " request_id=" + reqID
		}
	}

	// Add user ID if available from context
	if userID, ok := GetUserID(r); ok {
		msg += " user_id=" + userID
	}

	return msg
}

// RequestID middleware that adds a unique request ID to each request
func (m *LoggingMiddleware) RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request ID already exists
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// Generate new request ID
			requestID = generateRequestID()
		}

		// Add request ID to response headers
		w.Header().Set("X-Request-ID", requestID)

		// Add request ID to request headers for downstream handlers
		r.Header.Set("X-Request-ID", requestID)

		// Continue with request
		next.ServeHTTP(w, r)
	})
}

// generateRequestID generates a unique request ID
func generateRequestID() string {
	entropy := ulid.Monotonic(rand.New(rand.NewSource(time.Now().UnixNano())), 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy)
	return id.String()
}
