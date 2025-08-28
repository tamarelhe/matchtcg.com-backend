package middleware

import (
	"bytes"
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoggingMiddleware_Logging(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	t.Run("basic logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger: logger,
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String())

		logOutput := buf.String()
		assert.Contains(t, logOutput, "GET")
		assert.Contains(t, logOutput, "/test")
		assert.Contains(t, logOutput, "200")
		assert.Contains(t, logOutput, "7") // Response size
		assert.Contains(t, logOutput, "192.168.1.1")
	})

	t.Run("logging with query parameters", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger: logger,
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(testHandler)

		req := httptest.NewRequest("GET", "/test?param1=value1&param2=value2", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "query=param1=value1&param2=value2")
	})

	t.Run("logging with user agent", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger:       logger,
			LogUserAgent: true,
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("User-Agent", "TestAgent/1.0")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, `ua="TestAgent/1.0"`)
	})

	t.Run("logging with referer", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger:     logger,
			LogReferer: true,
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("Referer", "https://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, `referer="https://example.com"`)
	})

	t.Run("logging with request ID", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger:       logger,
			LogRequestID: true,
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("X-Request-ID", "req-123")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "request_id=req-123")
	})

	t.Run("logging with authenticated user", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger: logger,
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		ctx := context.WithValue(req.Context(), UserIDKey, "user123")
		req = req.WithContext(ctx)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "user_id=user123")
	})

	t.Run("skip paths", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger:    logger,
			SkipPaths: []string{"/health", "/metrics"},
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(testHandler)

		// Request to skipped path
		req := httptest.NewRequest("GET", "/health", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, buf.String()) // Should not log

		// Request to non-skipped path
		req = httptest.NewRequest("GET", "/api/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr = httptest.NewRecorder()
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.NotEmpty(t, buf.String()) // Should log
	})

	t.Run("error status code", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		errorHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("error"))
		})

		config := LoggingConfig{
			Logger: logger,
		}

		middleware := NewLoggingMiddleware(config)
		handler := middleware.Logging(errorHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		logOutput := buf.String()
		assert.Contains(t, logOutput, "500")
		assert.Contains(t, logOutput, "5") // Response size for "error"
	})
}

func TestLoggingMiddleware_RequestID(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("request_id:" + requestID))
	})

	middleware := NewLoggingMiddleware(LoggingConfig{})
	handler := middleware.RequestID(testHandler)

	t.Run("generates request ID when missing", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Check response header
		requestID := rr.Header().Get("X-Request-ID")
		assert.NotEmpty(t, requestID)

		// Check that it was passed to handler
		assert.Contains(t, rr.Body.String(), "request_id:"+requestID)
	})

	t.Run("preserves existing request ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "existing-123")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		// Check response header
		requestID := rr.Header().Get("X-Request-ID")
		assert.Equal(t, "existing-123", requestID)

		// Check that it was passed to handler
		assert.Contains(t, rr.Body.String(), "request_id:existing-123")
	})
}

func TestGenerateRequestID(t *testing.T) {
	id1 := generateRequestID()
	id2 := generateRequestID()

	assert.NotEmpty(t, id1)
	assert.NotEmpty(t, id2)
	assert.NotEqual(t, id1, id2) // Should be unique
}

func TestGetClientIPInLogging(t *testing.T) {
	// This function is already tested in rate_limit_middleware_test.go
	// but let's test it in the context of logging as well

	t.Run("X-Forwarded-For in logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := log.New(&buf, "", 0)

		config := LoggingConfig{
			Logger: logger,
		}

		middleware := NewLoggingMiddleware(config)
		testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})
		handler := middleware.Logging(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.1")
		req.RemoteAddr = "192.168.1.1:12345"
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		logOutput := buf.String()
		assert.Contains(t, logOutput, "203.0.113.1") // Should use first IP from X-Forwarded-For
	})
}
