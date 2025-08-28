package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/matchtcg/backend/internal/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	// Create JWT service for testing
	jwtService, err := service.NewJWTService(service.JWTConfig{})
	require.NoError(t, err)

	authMiddleware := NewAuthMiddleware(jwtService)

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r)
		assert.True(t, ok)
		assert.Equal(t, "test-user", userID)

		email, ok := GetEmail(r)
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", email)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Wrap with auth middleware
	handler := authMiddleware.RequireAuth(testHandler)

	t.Run("valid token", func(t *testing.T) {
		// Generate valid token
		tokenPair, err := jwtService.GenerateTokenPair("test-user", "test@example.com")
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "success", rr.Body.String())
	})

	t.Run("missing authorization header", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Authorization header required")
	})

	t.Run("invalid authorization format", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "InvalidFormat")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid authorization header format")
	})

	t.Run("empty token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer ")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Token required")
	})

	t.Run("invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid token")
	})

	t.Run("expired token", func(t *testing.T) {
		// Create service with very short TTL
		shortTTLService, err := service.NewJWTService(service.JWTConfig{
			AccessTTL: 1 * time.Millisecond,
		})
		require.NoError(t, err)

		tokenPair, err := shortTTLService.GenerateTokenPair("test-user", "test@example.com")
		require.NoError(t, err)

		// Wait for token to expire
		time.Sleep(10 * time.Millisecond)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		rr := httptest.NewRecorder()

		authMiddleware := NewAuthMiddleware(shortTTLService)
		handler := authMiddleware.RequireAuth(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Token expired")
	})

	t.Run("blacklisted token", func(t *testing.T) {
		blacklistStore := service.NewInMemoryBlacklistStore(time.Hour)
		defer blacklistStore.Close()

		blacklistService, err := service.NewJWTService(service.JWTConfig{
			BlacklistStore: blacklistStore,
		})
		require.NoError(t, err)

		tokenPair, err := blacklistService.GenerateTokenPair("test-user", "test@example.com")
		require.NoError(t, err)

		// Blacklist the token
		err = blacklistService.BlacklistToken(tokenPair.AccessToken)
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		rr := httptest.NewRecorder()

		authMiddleware := NewAuthMiddleware(blacklistService)
		handler := authMiddleware.RequireAuth(testHandler)
		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusUnauthorized, rr.Code)
		assert.Contains(t, rr.Body.String(), "Token revoked")
	})
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	// Create JWT service for testing
	jwtService, err := service.NewJWTService(service.JWTConfig{})
	require.NoError(t, err)

	authMiddleware := NewAuthMiddleware(jwtService)

	// Create a test handler that checks for user context
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, hasUser := GetUserID(r)
		if hasUser {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("authenticated: " + userID))
		} else {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("anonymous"))
		}
	})

	// Wrap with optional auth middleware
	handler := authMiddleware.OptionalAuth(testHandler)

	t.Run("with valid token", func(t *testing.T) {
		// Generate valid token
		tokenPair, err := jwtService.GenerateTokenPair("test-user", "test@example.com")
		require.NoError(t, err)

		req := httptest.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer "+tokenPair.AccessToken)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "authenticated: test-user", rr.Body.String())
	})

	t.Run("without token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/optional", nil)
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "anonymous", rr.Body.String())
	})

	t.Run("with invalid token", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/optional", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "anonymous", rr.Body.String())
	})
}

func TestContextHelpers(t *testing.T) {
	t.Run("GetUserID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		// Without user ID in context
		userID, ok := GetUserID(req)
		assert.False(t, ok)
		assert.Empty(t, userID)

		// With user ID in context
		ctx := context.WithValue(req.Context(), UserIDKey, "test-user")
		req = req.WithContext(ctx)

		userID, ok = GetUserID(req)
		assert.True(t, ok)
		assert.Equal(t, "test-user", userID)
	})

	t.Run("GetEmail", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		// Without email in context
		email, ok := GetEmail(req)
		assert.False(t, ok)
		assert.Empty(t, email)

		// With email in context
		ctx := context.WithValue(req.Context(), EmailKey, "test@example.com")
		req = req.WithContext(ctx)

		email, ok = GetEmail(req)
		assert.True(t, ok)
		assert.Equal(t, "test@example.com", email)
	})

	t.Run("GetClaims", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/test", nil)

		// Without claims in context
		claims, ok := GetClaims(req)
		assert.False(t, ok)
		assert.Nil(t, claims)

		// With claims in context
		testClaims := &service.TokenClaims{
			UserID: "test-user",
			Email:  "test@example.com",
		}
		ctx := context.WithValue(req.Context(), ClaimsKey, testClaims)
		req = req.WithContext(ctx)

		claims, ok = GetClaims(req)
		assert.True(t, ok)
		assert.Equal(t, testClaims, claims)
	})
}
