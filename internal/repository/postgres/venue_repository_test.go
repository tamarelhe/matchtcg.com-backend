package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestVenueRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test user
	user := createTestUser(t, db)

	venue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Test Venue",
		Type:      domain.VenueTypeStore,
		Address:   "123 Test Street",
		City:      "Lisbon",
		Country:   "PT",
		Latitude:  38.7223,
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{"capacity": 50, "parking": true},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, venue)
	require.NoError(t, err)

	// Verify venue was created
	retrieved, err := repo.GetByID(ctx, venue.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, venue.Name, retrieved.Name)
	assert.Equal(t, venue.Type, retrieved.Type)
	assert.Equal(t, venue.Address, retrieved.Address)
	assert.Equal(t, venue.City, retrieved.City)
	assert.Equal(t, venue.Country, retrieved.Country)
	assert.Equal(t, venue.Latitude, retrieved.Latitude)
	assert.Equal(t, venue.Longitude, retrieved.Longitude)
	assert.Equal(t, venue.Metadata, retrieved.Metadata)
	require.NotNil(t, retrieved.CreatedBy)
	assert.Equal(t, user.ID, *retrieved.CreatedBy)
}

func TestVenueRepository_SearchNearby(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test user
	user := createTestUser(t, db)

	// Create venues at different locations
	lisbonVenue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Lisbon Venue",
		Type:      domain.VenueTypeStore,
		Address:   "Rua Augusta, Lisbon",
		City:      "Lisbon",
		Country:   "PT",
		Latitude:  38.7223,
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	portoVenue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Porto Venue",
		Type:      domain.VenueTypeStore,
		Address:   "Rua de Santa Catarina, Porto",
		City:      "Porto",
		Country:   "PT",
		Latitude:  41.1579,
		Longitude: -8.6291,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, lisbonVenue)
	require.NoError(t, err)

	err = repo.Create(ctx, portoVenue)
	require.NoError(t, err)

	// Search near Lisbon (50km radius should include Lisbon venue but not Porto)
	nearbyVenues, err := repo.SearchNearby(ctx, 38.7223, -9.1393, 50, 10, 0)
	require.NoError(t, err)
	assert.Len(t, nearbyVenues, 1)
	assert.Equal(t, lisbonVenue.ID, nearbyVenues[0].ID)

	// Search with larger radius (400km should include both)
	nearbyVenues, err = repo.SearchNearby(ctx, 38.7223, -9.1393, 400, 10, 0)
	require.NoError(t, err)
	assert.Len(t, nearbyVenues, 2)
}

func TestVenueRepository_SearchByCity(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test user
	user := createTestUser(t, db)

	// Create venues in different cities
	lisbonVenue1 := createTestVenue(t, db, &user.ID)
	lisbonVenue1.City = "Lisbon"
	err := repo.Update(ctx, lisbonVenue1)
	require.NoError(t, err)

	lisbonVenue2 := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Another Lisbon Venue",
		Type:      domain.VenueTypeStore,
		Address:   "Another Address",
		City:      "Lisbon",
		Country:   "PT",
		Latitude:  38.7223,
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	err = repo.Create(ctx, lisbonVenue2)
	require.NoError(t, err)

	portoVenue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Porto Venue",
		Type:      domain.VenueTypeStore,
		Address:   "Porto Address",
		City:      "Porto",
		Country:   "PT",
		Latitude:  41.1579,
		Longitude: -8.6291,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	err = repo.Create(ctx, portoVenue)
	require.NoError(t, err)

	// Search by city
	lisbonVenues, err := repo.SearchByCity(ctx, "Lisbon", 10, 0)
	require.NoError(t, err)
	assert.Len(t, lisbonVenues, 2)

	portoVenues, err := repo.SearchByCity(ctx, "Porto", 10, 0)
	require.NoError(t, err)
	assert.Len(t, portoVenues, 1)
	assert.Equal(t, portoVenue.ID, portoVenues[0].ID)
}

func TestVenueRepository_SearchByName(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test user
	user := createTestUser(t, db)

	// Create venues with different names
	venue1 := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Magic Store Lisbon",
		Type:      domain.VenueTypeStore,
		Address:   "Address 1",
		City:      "Lisbon",
		Country:   "PT",
		Latitude:  38.7223,
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	venue2 := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Card Magic Shop",
		Type:      domain.VenueTypeStore,
		Address:   "Address 2",
		City:      "Porto",
		Country:   "PT",
		Latitude:  41.1579,
		Longitude: -8.6291,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	venue3 := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Board Game Cafe",
		Type:      domain.VenueTypeOther,
		Address:   "Address 3",
		City:      "Lisbon",
		Country:   "PT",
		Latitude:  38.7223,
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, venue1)
	require.NoError(t, err)

	err = repo.Create(ctx, venue2)
	require.NoError(t, err)

	err = repo.Create(ctx, venue3)
	require.NoError(t, err)

	// Search by name containing "Magic"
	magicVenues, err := repo.SearchByName(ctx, "Magic", 10, 0)
	require.NoError(t, err)
	assert.Len(t, magicVenues, 2)

	// Search by name containing "Board"
	boardVenues, err := repo.SearchByName(ctx, "Board", 10, 0)
	require.NoError(t, err)
	assert.Len(t, boardVenues, 1)
	assert.Equal(t, venue3.ID, boardVenues[0].ID)
}

func TestVenueRepository_FindNearestVenue(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test user
	user := createTestUser(t, db)

	// Create venues at different distances from a point
	nearVenue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Near Venue",
		Type:      domain.VenueTypeStore,
		Address:   "Near Address",
		City:      "Lisbon",
		Country:   "PT",
		Latitude:  38.7223, // Very close to search point
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	farVenue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Far Venue",
		Type:      domain.VenueTypeStore,
		Address:   "Far Address",
		City:      "Porto",
		Country:   "PT",
		Latitude:  41.1579, // Much farther
		Longitude: -8.6291,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	err := repo.Create(ctx, nearVenue)
	require.NoError(t, err)

	err = repo.Create(ctx, farVenue)
	require.NoError(t, err)

	// Find nearest venue to Lisbon coordinates
	nearest, err := repo.FindNearestVenue(ctx, 38.7223, -9.1393)
	require.NoError(t, err)
	require.NotNil(t, nearest)
	assert.Equal(t, nearVenue.ID, nearest.ID)
}

func TestVenueRepository_GetByType(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewVenueRepository(db)
	ctx := context.Background()

	// Create test user
	user := createTestUser(t, db)

	// Create venues of different types
	storeVenue := createTestVenue(t, db, &user.ID)
	storeVenue.Type = domain.VenueTypeStore
	err := repo.Update(ctx, storeVenue)
	require.NoError(t, err)

	homeVenue := &domain.Venue{
		ID:        uuid.New(),
		Name:      "Home Venue",
		Type:      domain.VenueTypeHome,
		Address:   "Home Address",
		City:      "Lisbon",
		Country:   "PT",
		Latitude:  38.7223,
		Longitude: -9.1393,
		Metadata:  map[string]interface{}{},
		CreatedBy: &user.ID,
		CreatedAt: time.Now(),
	}

	err = repo.Create(ctx, homeVenue)
	require.NoError(t, err)

	// Get venues by type
	storeVenues, err := repo.GetByType(ctx, domain.VenueTypeStore, 10, 0)
	require.NoError(t, err)
	assert.Len(t, storeVenues, 1)
	assert.Equal(t, storeVenue.ID, storeVenues[0].ID)

	homeVenues, err := repo.GetByType(ctx, domain.VenueTypeHome, 10, 0)
	require.NoError(t, err)
	assert.Len(t, homeVenues, 1)
	assert.Equal(t, homeVenue.ID, homeVenues[0].ID)
}
