package service

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryBlacklistStore_BlacklistToken(t *testing.T) {
	store := NewInMemoryBlacklistStore(time.Hour)
	defer store.Close()

	tokenID := "test-token-id"
	expiresAt := time.Now().Add(time.Hour)

	// Token should not be blacklisted initially
	blacklisted, err := store.IsBlacklisted(tokenID)
	require.NoError(t, err)
	assert.False(t, blacklisted)

	// Blacklist the token
	err = store.BlacklistToken(tokenID, expiresAt)
	require.NoError(t, err)

	// Token should now be blacklisted
	blacklisted, err = store.IsBlacklisted(tokenID)
	require.NoError(t, err)
	assert.True(t, blacklisted)
}

func TestInMemoryBlacklistStore_ExpiredToken(t *testing.T) {
	store := NewInMemoryBlacklistStore(time.Hour)
	defer store.Close()

	tokenID := "expired-token-id"
	expiresAt := time.Now().Add(-time.Hour) // Already expired

	// Blacklist an expired token
	err := store.BlacklistToken(tokenID, expiresAt)
	require.NoError(t, err)

	// Expired token should not be considered blacklisted
	blacklisted, err := store.IsBlacklisted(tokenID)
	require.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestInMemoryBlacklistStore_TokenExpiration(t *testing.T) {
	store := NewInMemoryBlacklistStore(time.Hour)
	defer store.Close()

	tokenID := "short-lived-token"
	expiresAt := time.Now().Add(50 * time.Millisecond)

	// Blacklist token with short expiration
	err := store.BlacklistToken(tokenID, expiresAt)
	require.NoError(t, err)

	// Token should be blacklisted initially
	blacklisted, err := store.IsBlacklisted(tokenID)
	require.NoError(t, err)
	assert.True(t, blacklisted)

	// Wait for token to expire
	time.Sleep(100 * time.Millisecond)

	// Token should no longer be blacklisted
	blacklisted, err = store.IsBlacklisted(tokenID)
	require.NoError(t, err)
	assert.False(t, blacklisted)
}

func TestInMemoryBlacklistStore_MultipleTokens(t *testing.T) {
	store := NewInMemoryBlacklistStore(time.Hour)
	defer store.Close()

	tokens := []struct {
		id        string
		expiresAt time.Time
	}{
		{"token1", time.Now().Add(time.Hour)},
		{"token2", time.Now().Add(2 * time.Hour)},
		{"token3", time.Now().Add(3 * time.Hour)},
	}

	// Blacklist all tokens
	for _, token := range tokens {
		err := store.BlacklistToken(token.id, token.expiresAt)
		require.NoError(t, err)
	}

	// All tokens should be blacklisted
	for _, token := range tokens {
		blacklisted, err := store.IsBlacklisted(token.id)
		require.NoError(t, err)
		assert.True(t, blacklisted, "token %s should be blacklisted", token.id)
	}

	// Check store size
	assert.Equal(t, 3, store.Size())
}

func TestInMemoryBlacklistStore_Cleanup(t *testing.T) {
	// Use very short cleanup interval for testing
	store := NewInMemoryBlacklistStore(10 * time.Millisecond)
	defer store.Close()

	// Add some tokens with different expiration times
	now := time.Now()
	tokens := []struct {
		id           string
		expiresAt    time.Time
		shouldExpire bool
	}{
		{"expired1", now.Add(-time.Hour), true},
		{"expired2", now.Add(-time.Minute), true},
		{"valid1", now.Add(time.Hour), false},
		{"valid2", now.Add(2 * time.Hour), false},
		{"short-lived", now.Add(50 * time.Millisecond), true},
	}

	for _, token := range tokens {
		err := store.BlacklistToken(token.id, token.expiresAt)
		require.NoError(t, err)
	}

	// Initially, all tokens should be in the store (even expired ones)
	assert.Equal(t, 5, store.Size())

	// Wait for cleanup to run and short-lived token to expire
	time.Sleep(100 * time.Millisecond)

	// Only valid tokens should remain
	expectedRemaining := 0
	for _, token := range tokens {
		if !token.shouldExpire {
			expectedRemaining++
		}
	}

	// Give cleanup some time to run
	time.Sleep(50 * time.Millisecond)

	// Check that expired tokens are cleaned up
	assert.LessOrEqual(t, store.Size(), expectedRemaining)

	// Valid tokens should still be blacklisted
	for _, token := range tokens {
		if !token.shouldExpire {
			blacklisted, err := store.IsBlacklisted(token.id)
			require.NoError(t, err)
			assert.True(t, blacklisted, "token %s should still be blacklisted", token.id)
		}
	}
}

func TestInMemoryBlacklistStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryBlacklistStore(time.Hour)
	defer store.Close()

	const numGoroutines = 10
	const tokensPerGoroutine = 100

	// Start multiple goroutines that add tokens concurrently
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < tokensPerGoroutine; j++ {
				tokenID := fmt.Sprintf("token-%d-%d", goroutineID, j)
				expiresAt := time.Now().Add(time.Hour)

				err := store.BlacklistToken(tokenID, expiresAt)
				require.NoError(t, err)

				// Immediately check if token is blacklisted
				blacklisted, err := store.IsBlacklisted(tokenID)
				require.NoError(t, err)
				assert.True(t, blacklisted)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final count
	expectedCount := numGoroutines * tokensPerGoroutine
	assert.Equal(t, expectedCount, store.Size())
}

func TestInMemoryBlacklistStore_Close(t *testing.T) {
	store := NewInMemoryBlacklistStore(10 * time.Millisecond)

	// Add a token
	err := store.BlacklistToken("test-token", time.Now().Add(time.Hour))
	require.NoError(t, err)

	// Close the store
	store.Close()

	// Store should still work for basic operations after close
	blacklisted, err := store.IsBlacklisted("test-token")
	require.NoError(t, err)
	assert.True(t, blacklisted)

	// But cleanup goroutine should have stopped
	// (This is hard to test directly, but Close() should not hang)
}

func TestInMemoryBlacklistStore_DefaultCleanupInterval(t *testing.T) {
	// Test with zero cleanup interval (should use default)
	store := NewInMemoryBlacklistStore(0)
	defer store.Close()

	// Should not panic and should work normally
	err := store.BlacklistToken("test-token", time.Now().Add(time.Hour))
	require.NoError(t, err)

	blacklisted, err := store.IsBlacklisted("test-token")
	require.NoError(t, err)
	assert.True(t, blacklisted)
}
