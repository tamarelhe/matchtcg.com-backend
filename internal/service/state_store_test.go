package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInMemoryStateStore_StorePKCEChallenge(t *testing.T) {
	store := NewInMemoryStateStore(10 * time.Minute)
	defer store.Close()

	state := "test-state"
	challenge := PKCEChallenge{
		CodeVerifier:  "test-verifier",
		CodeChallenge: "test-challenge",
		State:         state,
	}

	// Store challenge
	err := store.StorePKCEChallenge(state, challenge)
	require.NoError(t, err)

	// Retrieve challenge
	retrieved, err := store.GetPKCEChallenge(state)
	require.NoError(t, err)
	assert.Equal(t, challenge.CodeVerifier, retrieved.CodeVerifier)
	assert.Equal(t, challenge.CodeChallenge, retrieved.CodeChallenge)
	assert.Equal(t, challenge.State, retrieved.State)
	assert.True(t, retrieved.CreatedAt.After(time.Time{}))
}

func TestInMemoryStateStore_GetPKCEChallenge(t *testing.T) {
	store := NewInMemoryStateStore(10 * time.Minute)
	defer store.Close()

	t.Run("existing challenge", func(t *testing.T) {
		state := "existing-state"
		challenge := PKCEChallenge{
			CodeVerifier:  "verifier",
			CodeChallenge: "challenge",
			State:         state,
		}

		err := store.StorePKCEChallenge(state, challenge)
		require.NoError(t, err)

		retrieved, err := store.GetPKCEChallenge(state)
		require.NoError(t, err)
		assert.Equal(t, challenge.CodeVerifier, retrieved.CodeVerifier)
	})

	t.Run("non-existing challenge", func(t *testing.T) {
		_, err := store.GetPKCEChallenge("non-existing-state")
		assert.ErrorIs(t, err, ErrStateNotFound)
	})

	t.Run("expired challenge", func(t *testing.T) {
		// Create store with very short TTL
		shortStore := NewInMemoryStateStore(1 * time.Millisecond)
		defer shortStore.Close()

		state := "expired-state"
		challenge := PKCEChallenge{
			CodeVerifier:  "verifier",
			CodeChallenge: "challenge",
			State:         state,
		}

		err := shortStore.StorePKCEChallenge(state, challenge)
		require.NoError(t, err)

		// Wait for expiration
		time.Sleep(10 * time.Millisecond)

		_, err = shortStore.GetPKCEChallenge(state)

		require.True(t,
			errors.Is(err, ErrStateExpired) || errors.Is(err, ErrStateNotFound),
			"expected ErrStateExpired or ErrStateNotFound, got %v", err,
		)
	})
}

func TestInMemoryStateStore_DeletePKCEChallenge(t *testing.T) {
	store := NewInMemoryStateStore(10 * time.Minute)
	defer store.Close()

	state := "test-state"
	challenge := PKCEChallenge{
		CodeVerifier:  "verifier",
		CodeChallenge: "challenge",
		State:         state,
	}

	// Store challenge
	err := store.StorePKCEChallenge(state, challenge)
	require.NoError(t, err)

	// Verify it exists
	_, err = store.GetPKCEChallenge(state)
	require.NoError(t, err)

	// Delete challenge
	err = store.DeletePKCEChallenge(state)
	require.NoError(t, err)

	// Verify it's gone
	_, err = store.GetPKCEChallenge(state)
	assert.ErrorIs(t, err, ErrStateNotFound)
}

/* TODO
func TestInMemoryStateStore_Cleanup(t *testing.T) {
	// Use very short cleanup interval for testing
	store := NewInMemoryStateStore(10 * time.Millisecond)
	defer store.Close()

	// Add challenges with different creation times
	now := time.Now()
	challenges := []struct {
		state        string
		createdAt    time.Time
		shouldExpire bool
	}{
		{"old1", now.Add(-time.Hour), true},
		{"old2", now.Add(-30 * time.Minute), true},
		{"recent1", now.Add(-1 * time.Minute), false},
		{"recent2", now, false},
	}

	for _, c := range challenges {
		challenge := PKCEChallenge{
			CodeVerifier:  "verifier-" + c.state,
			CodeChallenge: "challenge-" + c.state,
			State:         c.state,
			CreatedAt:     c.createdAt,
		}

		// Manually set the creation time by storing in the map directly
		store.challenges[c.state] = challenge
	}

	// Initially, all challenges should be in the store
	assert.Equal(t, 4, store.Size())

	// Wait for cleanup to run
	time.Sleep(50 * time.Millisecond)

	// Check which challenges remain
	for _, c := range challenges {
		_, err := store.GetPKCEChallenge(c.state)
		if c.shouldExpire {
			assert.Error(t, err, "challenge %s should have been cleaned up", c.state)
		} else {
			assert.NoError(t, err, "challenge %s should still exist", c.state)
		}
	}
}
*/

func TestInMemoryStateStore_ConcurrentAccess(t *testing.T) {
	store := NewInMemoryStateStore(10 * time.Minute)
	defer store.Close()

	const numGoroutines = 10
	const challengesPerGoroutine = 100

	// Start multiple goroutines that store challenges concurrently
	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(goroutineID int) {
			defer func() { done <- true }()

			for j := 0; j < challengesPerGoroutine; j++ {
				state := fmt.Sprintf("state-%d-%d", goroutineID, j)
				challenge := PKCEChallenge{
					CodeVerifier:  fmt.Sprintf("verifier-%d-%d", goroutineID, j),
					CodeChallenge: fmt.Sprintf("challenge-%d-%d", goroutineID, j),
					State:         state,
				}

				err := store.StorePKCEChallenge(state, challenge)
				require.NoError(t, err)

				// Immediately try to retrieve
				retrieved, err := store.GetPKCEChallenge(state)
				require.NoError(t, err)
				assert.Equal(t, challenge.CodeVerifier, retrieved.CodeVerifier)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Verify final count
	expectedCount := numGoroutines * challengesPerGoroutine
	assert.Equal(t, expectedCount, store.Size())
}

func TestInMemoryStateStore_DefaultTTL(t *testing.T) {
	// Test with zero TTL (should use default)
	store := NewInMemoryStateStore(0)
	defer store.Close()

	// Should not panic and should work normally
	challenge := PKCEChallenge{
		CodeVerifier:  "verifier",
		CodeChallenge: "challenge",
		State:         "test-state",
	}

	err := store.StorePKCEChallenge("test-state", challenge)
	require.NoError(t, err)

	retrieved, err := store.GetPKCEChallenge("test-state")
	require.NoError(t, err)
	assert.Equal(t, challenge.CodeVerifier, retrieved.CodeVerifier)
}

func TestInMemoryStateStore_Close(t *testing.T) {
	store := NewInMemoryStateStore(10 * time.Millisecond)

	// Add a challenge
	challenge := PKCEChallenge{
		CodeVerifier:  "verifier",
		CodeChallenge: "challenge",
		State:         "test-state",
	}

	err := store.StorePKCEChallenge("test-state", challenge)
	require.NoError(t, err)

	// Close the store
	store.Close()

	// Store should still work for basic operations after close
	retrieved, err := store.GetPKCEChallenge("test-state")
	require.NoError(t, err)
	assert.Equal(t, challenge.CodeVerifier, retrieved.CodeVerifier)

	// But cleanup goroutine should have stopped
	// (This is hard to test directly, but Close() should not hang)
}

func TestInMemoryStateStore_MultipleStatesForSameChallenge(t *testing.T) {
	store := NewInMemoryStateStore(10 * time.Minute)
	defer store.Close()

	// Store multiple challenges
	challenges := map[string]PKCEChallenge{
		"state1": {
			CodeVerifier:  "verifier1",
			CodeChallenge: "challenge1",
			State:         "state1",
		},
		"state2": {
			CodeVerifier:  "verifier2",
			CodeChallenge: "challenge2",
			State:         "state2",
		},
		"state3": {
			CodeVerifier:  "verifier3",
			CodeChallenge: "challenge3",
			State:         "state3",
		},
	}

	// Store all challenges
	for state, challenge := range challenges {
		err := store.StorePKCEChallenge(state, challenge)
		require.NoError(t, err)
	}

	// Verify all can be retrieved
	for state, expected := range challenges {
		retrieved, err := store.GetPKCEChallenge(state)
		require.NoError(t, err)
		assert.Equal(t, expected.CodeVerifier, retrieved.CodeVerifier)
		assert.Equal(t, expected.CodeChallenge, retrieved.CodeChallenge)
		assert.Equal(t, expected.State, retrieved.State)
	}

	// Delete one challenge
	err := store.DeletePKCEChallenge("state2")
	require.NoError(t, err)

	// Verify it's gone but others remain
	_, err = store.GetPKCEChallenge("state2")
	assert.ErrorIs(t, err, ErrStateNotFound)

	_, err = store.GetPKCEChallenge("state1")
	assert.NoError(t, err)

	_, err = store.GetPKCEChallenge("state3")
	assert.NoError(t, err)

	assert.Equal(t, 2, store.Size())
}
