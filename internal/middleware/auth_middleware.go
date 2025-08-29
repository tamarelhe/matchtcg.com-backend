package middleware

import (
	"context"
	"net"
	"net/http"
	"strings"

	"github.com/matchtcg/backend/internal/service"
)

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	jwtService *service.JWTService
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(jwtService *service.JWTService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService: jwtService,
	}
}

// contextKey is used for context keys to avoid collisions
type contextKey string

const (
	UserIDKey contextKey = "user_id"
	EmailKey  contextKey = "email"
	ClaimsKey contextKey = "claims"
)

// RequireAuth middleware that requires valid JWT authentication
func (m *AuthMiddleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check for Bearer token format
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if token == "" {
			http.Error(w, "Token required", http.StatusUnauthorized)
			return
		}

		// Validate token
		claims, err := m.jwtService.ValidateAccessToken(token)
		if err != nil {
			switch err {
			case service.ErrTokenExpired:
				http.Error(w, "Token expired", http.StatusUnauthorized)
			case service.ErrTokenBlacklisted:
				http.Error(w, "Token revoked", http.StatusUnauthorized)
			default:
				http.Error(w, "Invalid token", http.StatusUnauthorized)
			}
			return
		}

		// Add claims to request context
		ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
		ctx = context.WithValue(ctx, EmailKey, claims.Email)
		ctx = context.WithValue(ctx, ClaimsKey, claims)

		// Continue with authenticated request
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware that extracts user info if token is present but doesn't require it
func (m *AuthMiddleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" && parts[1] != "" {
				// Validate token
				claims, err := m.jwtService.ValidateAccessToken(parts[1])
				if err == nil {
					// Add claims to request context
					ctx := context.WithValue(r.Context(), UserIDKey, claims.UserID)
					ctx = context.WithValue(ctx, EmailKey, claims.Email)
					ctx = context.WithValue(ctx, ClaimsKey, claims)
					r = r.WithContext(ctx)
				}
			}
		}

		// Continue with request (authenticated or not)
		next.ServeHTTP(w, r)
	})
}

func (m *AuthMiddleware) AllowOnlyLocal(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if ip != "127.0.0.1" && ip != "::1" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// GetUserID extracts user ID from request context
func GetUserID(r *http.Request) (string, bool) {
	userID, ok := r.Context().Value(UserIDKey).(string)
	return userID, ok
}

// GetEmail extracts email from request context
func GetEmail(r *http.Request) (string, bool) {
	email, ok := r.Context().Value(EmailKey).(string)
	return email, ok
}

// GetClaims extracts JWT claims from request context
func GetClaims(r *http.Request) (*service.TokenClaims, bool) {
	claims, ok := r.Context().Value(ClaimsKey).(*service.TokenClaims)
	return claims, ok
}
