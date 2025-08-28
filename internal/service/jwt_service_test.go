package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateTokenPair(t *testing.T) {
	// Create JWT service with default config
	jwtService, err := NewJWTService(JWTConfig{})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	// Generate token pair
	tokenPair, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.True(t, tokenPair.ExpiresAt.After(time.Now()))

	// Validate access token
	accessClaims, err := jwtService.ValidateAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, accessClaims.UserID)
	assert.Equal(t, email, accessClaims.Email)
	assert.Equal(t, userID, accessClaims.Subject)
	assert.Contains(t, accessClaims.Audience, "matchtcg-app")

	// Validate refresh token
	refreshClaims, err := jwtService.ValidateRefreshToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	assert.Equal(t, userID, refreshClaims.UserID)
	assert.Equal(t, email, refreshClaims.Email)
	assert.Contains(t, refreshClaims.Audience, "matchtcg-refresh")
}

func TestJWTService_ValidateAccessToken(t *testing.T) {
	jwtService, err := NewJWTService(JWTConfig{})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	tokenPair, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	t.Run("valid token", func(t *testing.T) {
		claims, err := jwtService.ValidateAccessToken(tokenPair.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
	})

	t.Run("invalid token", func(t *testing.T) {
		_, err := jwtService.ValidateAccessToken("invalid-token")
		assert.ErrorIs(t, err, ErrInvalidToken)
	})

	t.Run("refresh token used as access token", func(t *testing.T) {
		_, err := jwtService.ValidateAccessToken(tokenPair.RefreshToken)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}

func TestJWTService_ValidateRefreshToken(t *testing.T) {
	jwtService, err := NewJWTService(JWTConfig{})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	tokenPair, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	t.Run("valid refresh token", func(t *testing.T) {
		claims, err := jwtService.ValidateRefreshToken(tokenPair.RefreshToken)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.Equal(t, email, claims.Email)
	})

	t.Run("access token used as refresh token", func(t *testing.T) {
		_, err := jwtService.ValidateRefreshToken(tokenPair.AccessToken)
		assert.ErrorIs(t, err, ErrInvalidToken)
	})
}

func TestJWTService_RefreshTokens(t *testing.T) {
	blacklistStore := NewInMemoryBlacklistStore(time.Hour)
	defer blacklistStore.Close()

	jwtService, err := NewJWTService(JWTConfig{
		BlacklistStore: blacklistStore,
	})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	// Generate initial token pair
	initialTokens, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Refresh tokens
	newTokens, err := jwtService.RefreshTokens(initialTokens.RefreshToken)
	require.NoError(t, err)
	assert.NotEmpty(t, newTokens.AccessToken)
	assert.NotEmpty(t, newTokens.RefreshToken)
	assert.NotEqual(t, initialTokens.AccessToken, newTokens.AccessToken)
	assert.NotEqual(t, initialTokens.RefreshToken, newTokens.RefreshToken)

	// Validate new tokens
	claims, err := jwtService.ValidateAccessToken(newTokens.AccessToken)
	require.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)

	// Old refresh token should be blacklisted
	_, err = jwtService.RefreshTokens(initialTokens.RefreshToken)
	assert.ErrorIs(t, err, ErrTokenBlacklisted)
}

func TestJWTService_BlacklistToken(t *testing.T) {
	blacklistStore := NewInMemoryBlacklistStore(time.Hour)
	defer blacklistStore.Close()

	jwtService, err := NewJWTService(JWTConfig{
		BlacklistStore: blacklistStore,
	})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	tokenPair, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Token should be valid initially
	_, err = jwtService.ValidateAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)

	// Blacklist the token
	err = jwtService.BlacklistToken(tokenPair.AccessToken)
	require.NoError(t, err)

	// Token should now be invalid
	_, err = jwtService.ValidateAccessToken(tokenPair.AccessToken)
	assert.ErrorIs(t, err, ErrTokenBlacklisted)
}

func TestJWTService_ExpiredToken(t *testing.T) {
	// Create service with very short TTL
	jwtService, err := NewJWTService(JWTConfig{
		AccessTTL:  1 * time.Millisecond,
		RefreshTTL: 1 * time.Millisecond,
	})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	tokenPair, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	// Token should be expired
	_, err = jwtService.ValidateAccessToken(tokenPair.AccessToken)
	assert.ErrorIs(t, err, ErrTokenExpired)

	_, err = jwtService.ValidateRefreshToken(tokenPair.RefreshToken)
	assert.ErrorIs(t, err, ErrTokenExpired)
}

func TestJWTService_CustomTTL(t *testing.T) {
	accessTTL := 30 * time.Minute
	refreshTTL := 14 * 24 * time.Hour

	jwtService, err := NewJWTService(JWTConfig{
		AccessTTL:  accessTTL,
		RefreshTTL: refreshTTL,
	})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	tokenPair, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Validate access token claims
	accessClaims, err := jwtService.ValidateAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)

	expectedAccessExpiry := time.Now().Add(accessTTL)
	assert.WithinDuration(t, expectedAccessExpiry, accessClaims.ExpiresAt.Time, 5*time.Second)

	// Validate refresh token claims
	refreshClaims, err := jwtService.ValidateRefreshToken(tokenPair.RefreshToken)
	require.NoError(t, err)

	expectedRefreshExpiry := time.Now().Add(refreshTTL)
	assert.WithinDuration(t, expectedRefreshExpiry, refreshClaims.ExpiresAt.Time, 5*time.Second)
}

func TestJWTService_WithoutBlacklistStore(t *testing.T) {
	// Create service without blacklist store
	jwtService, err := NewJWTService(JWTConfig{})
	require.NoError(t, err)

	userID := "test-user-id"
	email := "test@example.com"

	tokenPair, err := jwtService.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Blacklisting should not error but also not affect validation
	err = jwtService.BlacklistToken(tokenPair.AccessToken)
	require.NoError(t, err)

	// Token should still be valid (no blacklist store to check)
	_, err = jwtService.ValidateAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)
}

func TestJWTService_InvalidSigningMethod(t *testing.T) {
	jwtService, err := NewJWTService(JWTConfig{})
	require.NoError(t, err)

	// Create a token with wrong signing method (HS256 instead of RS256)
	// This would be an attack attempt
	invalidToken := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoidGVzdCIsImVtYWlsIjoidGVzdEBleGFtcGxlLmNvbSIsImF1ZCI6WyJtYXRjaHRjZy1hcHAiXSwiZXhwIjo5OTk5OTk5OTk5LCJpYXQiOjE2MDAwMDAwMDAsImlzcyI6Im1hdGNodGNnLWJhY2tlbmQiLCJzdWIiOiJ0ZXN0In0.invalid"

	_, err = jwtService.ValidateAccessToken(invalidToken)
	assert.ErrorIs(t, err, ErrInvalidToken)
}
