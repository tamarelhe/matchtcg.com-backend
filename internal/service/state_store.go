package service

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrStateNotFound = errors.New("state not found")
	ErrStateExpired  = errors.New("state expired")
)

// InMemoryStateStore is a simple in-memory implementation of StateStore
// This is suitable for development and single-instance deployments
// For production with multiple instances, use Redis or database-backed store
type InMemoryStateStore struct {
	mu         sync.RWMutex
	challenges map[string]PKCEChallenge
	ttl        time.Duration
	stopCh     chan struct{}
}

// NewInMemoryStateStore creates a new in-memory state store
func NewInMemoryStateStore(ttl time.Duration) *InMemoryStateStore {
	if ttl == 0 {
		ttl = 10 * time.Minute // Default TTL for OAuth state
	}

	store := &InMemoryStateStore{
		challenges: make(map[string]PKCEChallenge),
		ttl:        ttl,
		stopCh:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// StorePKCEChallenge stores a PKCE challenge with the given state
func (s *InMemoryStateStore) StorePKCEChallenge(state string, challenge PKCEChallenge) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Set expiration time
	challenge.CreatedAt = time.Now()
	s.challenges[state] = challenge

	return nil
}

// GetPKCEChallenge retrieves a PKCE challenge by state
func (s *InMemoryStateStore) GetPKCEChallenge(state string) (*PKCEChallenge, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	challenge, exists := s.challenges[state]
	if !exists {
		return nil, ErrStateNotFound
	}

	// Check if expired
	if time.Since(challenge.CreatedAt) > s.ttl {
		// Remove expired challenge in a separate goroutine to avoid blocking
		go func() {
			s.mu.Lock()
			delete(s.challenges, state)
			s.mu.Unlock()
		}()
		return nil, ErrStateExpired
	}

	return &challenge, nil
}

// DeletePKCEChallenge removes a PKCE challenge by state
func (s *InMemoryStateStore) DeletePKCEChallenge(state string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.challenges, state)
	return nil
}

// Close stops the cleanup goroutine
func (s *InMemoryStateStore) Close() {
	close(s.stopCh)
}

// cleanupExpired removes expired challenges from the store
func (s *InMemoryStateStore) cleanupExpired() {
	ticker := time.NewTicker(s.ttl / 2) // Cleanup twice per TTL period
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.performCleanup()
		case <-s.stopCh:
			return
		}
	}
}

// performCleanup removes expired entries from the store
func (s *InMemoryStateStore) performCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for state, challenge := range s.challenges {
		if now.Sub(challenge.CreatedAt) > s.ttl {
			delete(s.challenges, state)
		}
	}
}

// Size returns the current number of stored challenges (for testing/monitoring)
func (s *InMemoryStateStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.challenges)
}
