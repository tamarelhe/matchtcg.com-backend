package middleware

import (
	"net/http"
	"strconv"
	"strings"
)

// CORSConfig holds configuration for CORS middleware
type CORSConfig struct {
	AllowedOrigins     []string // Allowed origins (use ["*"] for all)
	AllowedMethods     []string // Allowed HTTP methods
	AllowedHeaders     []string // Allowed headers
	ExposedHeaders     []string // Headers exposed to the client
	AllowCredentials   bool     // Whether to allow credentials
	MaxAge             int      // Preflight cache duration in seconds
	OptionsPassthrough bool     // Whether to pass OPTIONS requests to next handler
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},
		ExposedHeaders: []string{
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
		},
		AllowCredentials:   false,
		MaxAge:             86400, // 24 hours
		OptionsPassthrough: false,
	}
}

// CORSMiddleware handles Cross-Origin Resource Sharing
type CORSMiddleware struct {
	config CORSConfig
}

// NewCORSMiddleware creates a new CORS middleware
func NewCORSMiddleware(config CORSConfig) *CORSMiddleware {
	return &CORSMiddleware{
		config: config,
	}
}

// CORS middleware that handles CORS headers
func (m *CORSMiddleware) CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		isOriginAllowed, originMatchCfg := m.isOriginAllowed(origin)

		if isOriginAllowed {
			w.Header().Set("Access-Control-Allow-Origin", originMatchCfg)
		}

		// Set other CORS headers
		if len(m.config.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(m.config.AllowedMethods, ", "))
		}

		if len(m.config.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(m.config.AllowedHeaders, ", "))
		}

		if len(m.config.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(m.config.ExposedHeaders, ", "))
		}

		if m.config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if m.config.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", strconv.Itoa(m.config.MaxAge))
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			if !m.config.OptionsPassthrough {
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}

		// Continue with request
		next.ServeHTTP(w, r)
	})
}

// isOriginAllowed checks if the given origin is allowed
func (m *CORSMiddleware) isOriginAllowed(origin string) (bool, string) {
	if origin == "" {
		return false, ""
	}

	for _, allowedOrigin := range m.config.AllowedOrigins {
		if allowedOrigin == "*" {
			return true, "*"
		}
	}

	for _, allowedOrigin := range m.config.AllowedOrigins {
		if allowedOrigin == origin {
			return true, origin
		}
		// Support wildcard subdomains (e.g., "*.example.com")
		if strings.HasPrefix(allowedOrigin, "*.") {
			domain := allowedOrigin[2:]
			if strings.HasSuffix(origin, "."+domain) || origin == domain {
				return true, origin
			}
		}
	}

	return false, ""
}
