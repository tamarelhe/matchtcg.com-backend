package middleware

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSecurityMiddleware_SecurityHeaders(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	t.Run("default security headers", func(t *testing.T) {
		config := DefaultSecurityConfig()
		middleware := NewSecurityMiddleware(config)
		handler := middleware.SecurityHeaders(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String())

		// Check security headers
		assert.Contains(t, rr.Header().Get("Content-Security-Policy"), "default-src 'self'")
		assert.Equal(t, "DENY", rr.Header().Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", rr.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "1; mode=block", rr.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", rr.Header().Get("Referrer-Policy"))
		assert.Contains(t, rr.Header().Get("Permissions-Policy"), "camera=()")
		assert.Equal(t, "", rr.Header().Get("Server")) // Should be empty
		assert.Equal(t, "noindex, nofollow, nosnippet, noarchive", rr.Header().Get("X-Robots-Tag"))

		// HSTS should not be set for HTTP
		assert.Empty(t, rr.Header().Get("Strict-Transport-Security"))
	})

	t.Run("HSTS header for HTTPS", func(t *testing.T) {
		config := DefaultSecurityConfig()
		middleware := NewSecurityMiddleware(config)
		handler := middleware.SecurityHeaders(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.TLS = &tls.ConnectionState{} // Simulate HTTPS
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "max-age=31536000; includeSubDomains; preload", rr.Header().Get("Strict-Transport-Security"))
	})

	t.Run("custom security config", func(t *testing.T) {
		config := SecurityConfig{
			ContentSecurityPolicy: "default-src 'none'",
			FrameOptions:          "SAMEORIGIN",
			ContentTypeOptions:    "nosniff",
			XSSProtection:         "0",
			ReferrerPolicy:        "no-referrer",
			CustomHeaders: map[string]string{
				"X-Custom-Header": "custom-value",
				"X-API-Version":   "v1.0",
			},
		}

		middleware := NewSecurityMiddleware(config)
		handler := middleware.SecurityHeaders(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "default-src 'none'", rr.Header().Get("Content-Security-Policy"))
		assert.Equal(t, "SAMEORIGIN", rr.Header().Get("X-Frame-Options"))
		assert.Equal(t, "0", rr.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "no-referrer", rr.Header().Get("Referrer-Policy"))
		assert.Equal(t, "custom-value", rr.Header().Get("X-Custom-Header"))
		assert.Equal(t, "v1.0", rr.Header().Get("X-API-Version"))
	})

	t.Run("empty config values", func(t *testing.T) {
		config := SecurityConfig{
			// All empty - should not set headers
		}

		middleware := NewSecurityMiddleware(config)
		handler := middleware.SecurityHeaders(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Content-Security-Policy"))
		assert.Empty(t, rr.Header().Get("X-Frame-Options"))
		assert.Empty(t, rr.Header().Get("X-Content-Type-Options"))
		assert.Empty(t, rr.Header().Get("X-XSS-Protection"))
		assert.Empty(t, rr.Header().Get("Referrer-Policy"))
		assert.Empty(t, rr.Header().Get("Permissions-Policy"))

		// Server header should still be set to empty
		assert.Equal(t, "", rr.Header().Get("Server"))
	})
}

func TestDefaultSecurityConfig(t *testing.T) {
	config := DefaultSecurityConfig()

	assert.Contains(t, config.ContentSecurityPolicy, "default-src 'self'")
	assert.Equal(t, "DENY", config.FrameOptions)
	assert.Equal(t, "nosniff", config.ContentTypeOptions)
	assert.Equal(t, "1; mode=block", config.XSSProtection)
	assert.Contains(t, config.StrictTransportSecurity, "max-age=31536000")
	assert.Equal(t, "strict-origin-when-cross-origin", config.ReferrerPolicy)
	assert.Contains(t, config.PermissionsPolicy, "camera=()")
	assert.NotEmpty(t, config.CustomHeaders)
	assert.Equal(t, "noindex, nofollow, nosnippet, noarchive", config.CustomHeaders["X-Robots-Tag"])
}
