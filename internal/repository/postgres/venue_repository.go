package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
	"github.com/matchtcg/backend/internal/service"
)

type venueRepository struct {
	db service.DB
}

// NewVenueRepository creates a new PostgreSQL venue repository
func NewVenueRepository(db service.DB) repository.VenueRepository {
	return &venueRepository{db: db}
}

// Create creates a new venue
func (r *venueRepository) Create(ctx context.Context, venue *domain.Venue) error {
	metadataJSON, err := json.Marshal(venue.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO venues (id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err = r.db.Exec(ctx, query,
		venue.ID,
		venue.Name,
		venue.Type,
		venue.Address,
		venue.City,
		venue.Country,
		venue.Latitude,
		venue.Longitude,
		metadataJSON,
		venue.CreatedBy,
		venue.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create venue: %w", err)
	}

	return nil
}

// GetByID retrieves a venue by ID
func (r *venueRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE id = $1`

	return r.scanVenue(r.db.QueryRow(ctx, query, id))
}

// Update updates a venue
func (r *venueRepository) Update(ctx context.Context, venue *domain.Venue) error {
	metadataJSON, err := json.Marshal(venue.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE venues
		SET name = $2, type = $3, address = $4, city = $5, country = $6,
			latitude = $7, longitude = $8, metadata = $9
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		venue.ID,
		venue.Name,
		venue.Type,
		venue.Address,
		venue.City,
		venue.Country,
		venue.Latitude,
		venue.Longitude,
		metadataJSON,
	)

	if err != nil {
		return fmt.Errorf("failed to update venue: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}

// Delete deletes a venue
func (r *venueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM venues WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete venue: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("venue not found")
	}

	return nil
}

// SearchNearby searches for venues near a location using PostGIS ST_DWithin
func (r *venueRepository) SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, limit, offset int) ([]*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE ST_DWithin(coordinates, ST_Point($1, $2)::geography, $3)
		ORDER BY ST_Distance(coordinates, ST_Point($1, $2)::geography)
		LIMIT $4 OFFSET $5`

	rows, err := r.db.Query(ctx, query, lon, lat, radiusKm*1000, limit, offset) // Convert km to meters
	if err != nil {
		return nil, fmt.Errorf("failed to search nearby venues: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// SearchByCity searches for venues in a specific city
func (r *venueRepository) SearchByCity(ctx context.Context, city string, limit, offset int) ([]*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE LOWER(city) = LOWER($1)
		ORDER BY name
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, city, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search venues by city: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// SearchByCountry searches for venues in a specific country
func (r *venueRepository) SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE LOWER(country) = LOWER($1)
		ORDER BY city, name
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, country, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search venues by country: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// SearchByName searches for venues by name (case-insensitive partial match)
func (r *venueRepository) SearchByName(ctx context.Context, name string, limit, offset int) ([]*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE LOWER(name) LIKE LOWER($1)
		ORDER BY name
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, "%"+name+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search venues by name: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// GetByCreator retrieves venues created by a specific user
func (r *venueRepository) GetByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE created_by = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, creatorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get venues by creator: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// GetByType retrieves venues of a specific type
func (r *venueRepository) GetByType(ctx context.Context, venueType domain.VenueType, limit, offset int) ([]*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE type = $1
		ORDER BY name
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, venueType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get venues by type: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// GetPopularVenues retrieves popular venues (based on event count)
func (r *venueRepository) GetPopularVenues(ctx context.Context, limit, offset int) ([]*domain.Venue, error) {
	query := `
		SELECT v.id, v.name, v.type, v.address, v.city, v.country, v.latitude, v.longitude, v.metadata, v.created_by, v.created_at
		FROM venues v
		LEFT JOIN events e ON v.id = e.venue_id
		GROUP BY v.id, v.name, v.type, v.address, v.city, v.country, v.latitude, v.longitude, v.metadata, v.created_by, v.created_at
		ORDER BY COUNT(e.id) DESC, v.name
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular venues: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// GetVenuesInBounds retrieves venues within geographic bounds
func (r *venueRepository) GetVenuesInBounds(ctx context.Context, northLat, southLat, eastLon, westLon float64) ([]*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		WHERE latitude BETWEEN $1 AND $2
		AND longitude BETWEEN $3 AND $4
		ORDER BY name`

	rows, err := r.db.Query(ctx, query, southLat, northLat, westLon, eastLon)
	if err != nil {
		return nil, fmt.Errorf("failed to get venues in bounds: %w", err)
	}
	defer rows.Close()

	var venues []*domain.Venue
	for rows.Next() {
		venue, err := r.scanVenue(rows)
		if err != nil {
			return nil, err
		}
		venues = append(venues, venue)
	}

	return venues, nil
}

// FindNearestVenue finds the nearest venue to a location using PostGIS KNN
func (r *venueRepository) FindNearestVenue(ctx context.Context, lat, lon float64) (*domain.Venue, error) {
	query := `
		SELECT id, name, type, address, city, country, latitude, longitude, metadata, created_by, created_at
		FROM venues
		ORDER BY coordinates <-> ST_Point($1, $2)::geography
		LIMIT 1`

	return r.scanVenue(r.db.QueryRow(ctx, query, lon, lat))
}

// CountVenuesByCity counts venues in a specific city
func (r *venueRepository) CountVenuesByCity(ctx context.Context, city string) (int, error) {
	query := `SELECT COUNT(*) FROM venues WHERE LOWER(city) = LOWER($1)`

	var count int
	err := r.db.QueryRow(ctx, query, city).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count venues by city: %w", err)
	}

	return count, nil
}

// CountVenuesByCountry counts venues in a specific country
func (r *venueRepository) CountVenuesByCountry(ctx context.Context, country string) (int, error) {
	query := `SELECT COUNT(*) FROM venues WHERE LOWER(country) = LOWER($1)`

	var count int
	err := r.db.QueryRow(ctx, query, country).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count venues by country: %w", err)
	}

	return count, nil
}

// Helper function to scan a venue from a row
func (r *venueRepository) scanVenue(row pgx.Row) (*domain.Venue, error) {
	var venue domain.Venue
	var metadataJSON []byte

	err := row.Scan(
		&venue.ID,
		&venue.Name,
		&venue.Type,
		&venue.Address,
		&venue.City,
		&venue.Country,
		&venue.Latitude,
		&venue.Longitude,
		&metadataJSON,
		&venue.CreatedBy,
		&venue.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan venue: %w", err)
	}

	// Unmarshal metadata
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &venue.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	return &venue, nil
}
