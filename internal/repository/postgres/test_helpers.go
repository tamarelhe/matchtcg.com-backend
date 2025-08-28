package postgres

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matchtcg/backend/internal/domain"
)

// setupTestDB creates a test database connection
// In a real implementation, this would set up a test database
func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	// Skip tests if no database URL is provided
	dbURL := os.Getenv("TEST_DATABASE_URL")
	if dbURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping database integration tests")
	}

	ctx := context.Background()
	db, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Test the connection
	if err := db.Ping(ctx); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	return db
}

// cleanupTestDB cleans up the test database
func cleanupTestDB(t *testing.T, db *pgxpool.Pool) {
	t.Helper()

	ctx := context.Background()

	// Clean up test data in reverse order of dependencies
	tables := []string{
		"notifications",
		"event_rsvp",
		"events",
		"venues",
		"group_members",
		"groups",
		"profiles",
		"users",
	}

	for _, table := range tables {
		_, err := db.Exec(ctx, fmt.Sprintf("DELETE FROM %s", table))
		if err != nil {
			t.Logf("Warning: Failed to clean up table %s: %v", table, err)
		}
	}

	db.Close()
}

// createTestUser creates a test user for use in other repository tests
func createTestUser(t *testing.T, db *pgxpool.Pool) *domain.User {
	t.Helper()

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        fmt.Sprintf("test-%s@example.com", uuid.New().String()[:8]),
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return user
}

// createTestGroup creates a test group for use in other repository tests
func createTestGroup(t *testing.T, db *pgxpool.Pool, ownerID uuid.UUID) *domain.Group {
	t.Helper()

	repo := NewGroupRepository(db)
	ctx := context.Background()

	group := &domain.Group{
		ID:          uuid.New(),
		Name:        fmt.Sprintf("Test Group %s", uuid.New().String()[:8]),
		OwnerUserID: ownerID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	err := repo.Create(ctx, group)
	if err != nil {
		t.Fatalf("Failed to create test group: %v", err)
	}

	return group
}

// createTestVenue creates a test venue for use in other repository tests
func createTestVenue(t *testing.T, db *pgxpool.Pool, createdBy *uuid.UUID) *domain.Venue {
	t.Helper()

	repo := NewVenueRepository(db)
	ctx := context.Background()

	venue := &domain.Venue{
		ID:        uuid.New(),
		Name:      fmt.Sprintf("Test Venue %s", uuid.New().String()[:8]),
		Type:      domain.VenueTypeStore,
		Address:   "123 Test Street",
		City:      "Test City",
		Country:   "PT",
		Latitude:  38.7223, // Lisbon coordinates
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{"test": true},
		CreatedBy: createdBy,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, venue)
	if err != nil {
		t.Fatalf("Failed to create test venue: %v", err)
	}

	return venue
}
