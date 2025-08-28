package service

import (
	"sync"
	"time"
)

// InMemoryBlacklistStore is a simple in-memory implementation of BlacklistStore
// This is suitable for development and single-instance deployments
// For production with multiple instances, use Redis or database-backed store
type InMemoryBlacklistStore struct {
	mu         sync.RWMutex
	blacklist  map[string]time.Time
	cleanupTTL time.Duration
	stopCh     chan struct{}
}

// NewInMemoryBlacklistStore creates a new in-memory blacklist store
func NewInMemoryBlacklistStore(cleanupInterval time.Duration) *InMemoryBlacklistStore {
	if cleanupInterval == 0 {
		cleanupInterval = 1 * time.Hour // Default cleanup every hour
	}

	store := &InMemoryBlacklistStore{
		blacklist:  make(map[string]time.Time),
		cleanupTTL: cleanupInterval,
		stopCh:     make(chan struct{}),
	}

	// Start cleanup goroutine
	go store.cleanupExpired()

	return store
}

// IsBlacklisted checks if a token ID is blacklisted
func (s *InMemoryBlacklistStore) IsBlacklisted(tokenID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	expiresAt, exists := s.blacklist[tokenID]
	if !exists {
		return false, nil
	}

	// Check if token has expired (and should be removed)
	if time.Now().After(expiresAt) {
		// Token expired, remove it in a separate goroutine to avoid blocking
		go func() {
			s.mu.Lock()
			delete(s.blacklist, tokenID)
			s.mu.Unlock()
		}()
		return false, nil
	}

	return true, nil
}

// BlacklistToken adds a token ID to the blacklist
func (s *InMemoryBlacklistStore) BlacklistToken(tokenID string, expiresAt time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.blacklist[tokenID] = expiresAt
	return nil
}

// Close stops the cleanup goroutine
func (s *InMemoryBlacklistStore) Close() {
	close(s.stopCh)
}

// cleanupExpired removes expired tokens from the blacklist
func (s *InMemoryBlacklistStore) cleanupExpired() {
	ticker := time.NewTicker(s.cleanupTTL)
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

// performCleanup removes expired entries from the blacklist
func (s *InMemoryBlacklistStore) performCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for tokenID, expiresAt := range s.blacklist {
		if now.After(expiresAt) {
			delete(s.blacklist, tokenID)
		}
	}
}

// Size returns the current number of blacklisted tokens (for testing/monitoring)
func (s *InMemoryBlacklistStore) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.blacklist)
}
