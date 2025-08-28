package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_CORS(t *testing.T) {
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	t.Run("wildcard origin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*"},
			AllowedMethods: []string{"GET", "POST"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
	})

	t.Run("specific allowed origin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"https://example.com", "https://app.example.com"},
			AllowedMethods: []string{"GET", "POST", "PUT"},
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT", rr.Header().Get("Access-Control-Allow-Methods"))
	})

	t.Run("disallowed origin", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"https://example.com"},
			AllowedMethods: []string{"GET", "POST"},
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://malicious.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("wildcard subdomain", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"*.example.com"},
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		// Test subdomain
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://app.example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "https://app.example.com", rr.Header().Get("Access-Control-Allow-Origin"))

		// Test main domain
		req = httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rr = httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "", rr.Header().Get("Access-Control-Allow-Origin"))

		// Test non-matching domain
		req = httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://other.com")
		rr = httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
	})

	t.Run("credentials and exposed headers", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins:   []string{"https://example.com"},
			AllowCredentials: true,
			ExposedHeaders:   []string{"X-Total-Count", "X-Page-Count"},
			MaxAge:           3600,
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
		assert.Equal(t, "X-Total-Count, X-Page-Count", rr.Header().Get("Access-Control-Expose-Headers"))
		assert.Equal(t, "3600", rr.Header().Get("Access-Control-Max-Age"))
	})

	t.Run("preflight request", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"https://example.com"},
			AllowedMethods: []string{"GET", "POST", "PUT"},
			AllowedHeaders: []string{"Content-Type", "Authorization"},
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		req.Header.Set("Access-Control-Request-Method", "POST")
		req.Header.Set("Access-Control-Request-Headers", "Content-Type")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code)
		assert.Equal(t, "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
		assert.Empty(t, rr.Body.String()) // Should not call next handler
	})

	t.Run("preflight with passthrough", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins:     []string{"https://example.com"},
			OptionsPassthrough: true,
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		req := httptest.NewRequest("OPTIONS", "/test", nil)
		req.Header.Set("Origin", "https://example.com")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String()) // Should call next handler
	})

	t.Run("no origin header", func(t *testing.T) {
		config := CORSConfig{
			AllowedOrigins: []string{"https://example.com"},
		}

		middleware := NewCORSMiddleware(config)
		handler := middleware.CORS(testHandler)

		req := httptest.NewRequest("GET", "/test", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Empty(t, rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "success", rr.Body.String())
	})
}

func TestDefaultCORSConfig(t *testing.T) {
	config := DefaultCORSConfig()

	assert.Equal(t, []string{"*"}, config.AllowedOrigins)
	assert.Contains(t, config.AllowedMethods, "GET")
	assert.Contains(t, config.AllowedMethods, "POST")
	assert.Contains(t, config.AllowedMethods, "PUT")
	assert.Contains(t, config.AllowedMethods, "DELETE")
	assert.Contains(t, config.AllowedHeaders, "Authorization")
	assert.Contains(t, config.AllowedHeaders, "Content-Type")
	assert.False(t, config.AllowCredentials)
	assert.Equal(t, 86400, config.MaxAge)
	assert.False(t, config.OptionsPassthrough)
}

func TestCORSMiddleware_isOriginAllowed(t *testing.T) {
	middleware := NewCORSMiddleware(CORSConfig{
		AllowedOrigins: []string{
			"https://example.com",
			"*.subdomain.com",
			"http://localhost:3000",
		},
	})

	testCases := []struct {
		origin   string
		expected bool
	}{
		{"", false},
		{"https://example.com", true},
		{"https://other.com", false},
		{"https://app.subdomain.com", true},
		{"https://api.subdomain.com", true},
		{"https://subdomain.com", false},
		{"https://notsubdomain.com", false},
		{"http://localhost:3000", true},
		{"http://localhost:3001", false},
	}

	for _, tc := range testCases {
		t.Run(tc.origin, func(t *testing.T) {
			result, _ := middleware.isOriginAllowed(tc.origin)
			assert.Equal(t, tc.expected, result)
		})
	}
}
