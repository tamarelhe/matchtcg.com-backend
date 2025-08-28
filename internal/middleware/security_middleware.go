package middleware

import (
	"net/http"
)

// SecurityConfig holds configuration for security headers
type SecurityConfig struct {
	// Content Security Policy
	ContentSecurityPolicy string

	// X-Frame-Options
	FrameOptions string

	// X-Content-Type-Options
	ContentTypeOptions string

	// X-XSS-Protection
	XSSProtection string

	// Strict-Transport-Security
	StrictTransportSecurity string

	// Referrer-Policy
	ReferrerPolicy string

	// Permissions-Policy
	PermissionsPolicy string

	// Custom headers
	CustomHeaders map[string]string
}

// DefaultSecurityConfig returns a default security configuration
func DefaultSecurityConfig() SecurityConfig {
	return SecurityConfig{
		ContentSecurityPolicy:   "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'; media-src 'self'; object-src 'none'; child-src 'none'; worker-src 'none'; frame-ancestors 'none'; form-action 'self'; base-uri 'self'; manifest-src 'self';",
		FrameOptions:            "DENY",
		ContentTypeOptions:      "nosniff",
		XSSProtection:           "1; mode=block",
		StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
		ReferrerPolicy:          "strict-origin-when-cross-origin",
		PermissionsPolicy:       "camera=(), microphone=(), geolocation=(), interest-cohort=()",
		CustomHeaders: map[string]string{
			"X-Robots-Tag": "noindex, nofollow, nosnippet, noarchive",
		},
	}
}

// SecurityMiddleware handles security headers
type SecurityMiddleware struct {
	config SecurityConfig
}

// NewSecurityMiddleware creates a new security middleware
func NewSecurityMiddleware(config SecurityConfig) *SecurityMiddleware {
	return &SecurityMiddleware{
		config: config,
	}
}

// SecurityHeaders middleware that adds security headers to responses
func (m *SecurityMiddleware) SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Content Security Policy
		if m.config.ContentSecurityPolicy != "" {
			w.Header().Set("Content-Security-Policy", m.config.ContentSecurityPolicy)
		}

		// X-Frame-Options
		if m.config.FrameOptions != "" {
			w.Header().Set("X-Frame-Options", m.config.FrameOptions)
		}

		// X-Content-Type-Options
		if m.config.ContentTypeOptions != "" {
			w.Header().Set("X-Content-Type-Options", m.config.ContentTypeOptions)
		}

		// X-XSS-Protection
		if m.config.XSSProtection != "" {
			w.Header().Set("X-XSS-Protection", m.config.XSSProtection)
		}

		// Strict-Transport-Security (only for HTTPS)
		if m.config.StrictTransportSecurity != "" && r.TLS != nil {
			w.Header().Set("Strict-Transport-Security", m.config.StrictTransportSecurity)
		}

		// Referrer-Policy
		if m.config.ReferrerPolicy != "" {
			w.Header().Set("Referrer-Policy", m.config.ReferrerPolicy)
		}

		// Permissions-Policy
		if m.config.PermissionsPolicy != "" {
			w.Header().Set("Permissions-Policy", m.config.PermissionsPolicy)
		}

		// Custom headers
		for key, value := range m.config.CustomHeaders {
			w.Header().Set(key, value)
		}

		// Remove server information
		w.Header().Set("Server", "")

		// Continue with request
		next.ServeHTTP(w, r)
	})
}
