package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
)

type eventRepository struct {
	db *pgxpool.Pool
}

// NewEventRepository creates a new PostgreSQL event repository
func NewEventRepository(db *pgxpool.Pool) repository.EventRepository {
	return &eventRepository{db: db}
}

// Create creates a new event
func (r *eventRepository) Create(ctx context.Context, event *domain.Event) error {
	rulesJSON, err := json.Marshal(event.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	query := `
		INSERT INTO events (id, host_user_id, group_id, venue_id, title, description, game, format,
			rules, visibility, capacity, start_at, end_at, timezone, tags, entry_fee, language,
			is_recurring, recurrence_rule, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`

	_, err = r.db.Exec(ctx, query,
		event.ID,
		event.HostUserID,
		event.GroupID,
		event.VenueID,
		event.Title,
		event.Description,
		event.Game,
		event.Format,
		rulesJSON,
		event.Visibility,
		event.Capacity,
		event.StartAt,
		event.EndAt,
		event.Timezone,
		event.Tags,
		event.EntryFee,
		event.Language,
		event.IsRecurring,
		event.RecurrenceRule,
		event.CreatedAt,
		event.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create event: %w", err)
	}

	return nil
}

// GetByID retrieves an event by ID
func (r *eventRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error) {
	query := `
		SELECT id, host_user_id, group_id, venue_id, title, description, game, format,
			rules, visibility, capacity, start_at, end_at, timezone, tags, entry_fee, language,
			is_recurring, recurrence_rule, created_at, updated_at
		FROM events
		WHERE id = $1`

	return r.scanEvent(r.db.QueryRow(ctx, query, id))
}

// GetByIDWithDetails retrieves an event with details by ID
func (r *eventRepository) GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.EventWithDetails, error) {
	query := `
		SELECT e.id, e.host_user_id, e.group_id, e.venue_id, e.title, e.description, e.game, e.format,
			e.rules, e.visibility, e.capacity, e.start_at, e.end_at, e.timezone, e.tags, e.entry_fee, e.language,
			e.is_recurring, e.recurrence_rule, e.created_at, e.updated_at,
			u.id, u.email, u.password_hash, u.created_at, u.updated_at, u.is_active, u.last_login,
			v.id, v.name, v.type, v.address, v.city, v.country, v.latitude, v.longitude, v.metadata, v.created_by, v.created_at,
			g.id, g.name, g.description, g.owner_user_id, g.created_at, g.updated_at, g.is_active
		FROM events e
		LEFT JOIN users u ON e.host_user_id = u.id
		LEFT JOIN venues v ON e.venue_id = v.id
		LEFT JOIN groups g ON e.group_id = g.id
		WHERE e.id = $1`

	var event domain.Event
	var host domain.User
	var venue domain.Venue
	var group domain.Group
	var rulesJSON []byte
	var venueMetadataJSON []byte

	var hostID, venueID, groupID sql.NullString
	var hostEmail, hostPasswordHash sql.NullString
	var hostCreatedAt, hostUpdatedAt sql.NullTime
	var hostIsActive sql.NullBool
	var hostLastLogin sql.NullTime
	var venueName, venueType, venueAddress, venueCity, venueCountry sql.NullString
	var venueLatitude, venueLongitude sql.NullFloat64
	var venueCreatedBy sql.NullString
	var venueCreatedAt sql.NullTime
	var groupName, groupDescription sql.NullString
	var groupOwnerUserID sql.NullString
	var groupCreatedAt, groupUpdatedAt sql.NullTime
	var groupIsActive sql.NullBool

	err := r.db.QueryRow(ctx, query, id).Scan(
		&event.ID, &event.HostUserID, &event.GroupID, &event.VenueID, &event.Title, &event.Description,
		&event.Game, &event.Format, &rulesJSON, &event.Visibility, &event.Capacity, &event.StartAt,
		&event.EndAt, &event.Timezone, &event.Tags, &event.EntryFee, &event.Language,
		&event.IsRecurring, &event.RecurrenceRule, &event.CreatedAt, &event.UpdatedAt,
		&hostID, &hostEmail, &hostPasswordHash, &hostCreatedAt, &hostUpdatedAt, &hostIsActive, &hostLastLogin,
		&venueID, &venueName, &venueType, &venueAddress, &venueCity, &venueCountry, &venueLatitude, &venueLongitude,
		&venueMetadataJSON, &venueCreatedBy, &venueCreatedAt,
		&groupID, &groupName, &groupDescription, &groupOwnerUserID, &groupCreatedAt, &groupUpdatedAt, &groupIsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get event with details: %w", err)
	}

	// Unmarshal rules
	if err := json.Unmarshal(rulesJSON, &event.Rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	eventWithDetails := &domain.EventWithDetails{Event: event}

	// Set host if exists
	if hostID.Valid {
		host.ID = uuid.MustParse(hostID.String)
		host.Email = hostEmail.String
		host.PasswordHash = hostPasswordHash.String
		host.CreatedAt = hostCreatedAt.Time
		host.UpdatedAt = hostUpdatedAt.Time
		host.IsActive = hostIsActive.Bool
		if hostLastLogin.Valid {
			host.LastLogin = &hostLastLogin.Time
		}
		eventWithDetails.Host = &host
	}

	// Set venue if exists
	if venueID.Valid {
		venue.ID = uuid.MustParse(venueID.String)
		venue.Name = venueName.String
		venue.Type = domain.VenueType(venueType.String)
		venue.Address = venueAddress.String
		venue.City = venueCity.String
		venue.Country = venueCountry.String
		venue.Latitude = venueLatitude.Float64
		venue.Longitude = venueLongitude.Float64
		if venueCreatedBy.Valid {
			createdBy := uuid.MustParse(venueCreatedBy.String)
			venue.CreatedBy = &createdBy
		}
		venue.CreatedAt = venueCreatedAt.Time

		if venueMetadataJSON != nil {
			if err := json.Unmarshal(venueMetadataJSON, &venue.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal venue metadata: %w", err)
			}
		}
		eventWithDetails.Venue = &venue
	}

	// Set group if exists
	if groupID.Valid {
		group.ID = uuid.MustParse(groupID.String)
		group.Name = groupName.String
		if groupDescription.Valid {
			group.Description = &groupDescription.String
		}
		group.OwnerUserID = uuid.MustParse(groupOwnerUserID.String)
		group.CreatedAt = groupCreatedAt.Time
		group.UpdatedAt = groupUpdatedAt.Time
		group.IsActive = groupIsActive.Bool
		eventWithDetails.Group = &group
	}

	return eventWithDetails, nil
}

// Update updates an event
func (r *eventRepository) Update(ctx context.Context, event *domain.Event) error {
	rulesJSON, err := json.Marshal(event.Rules)
	if err != nil {
		return fmt.Errorf("failed to marshal rules: %w", err)
	}

	query := `
		UPDATE events
		SET host_user_id = $2, group_id = $3, venue_id = $4, title = $5, description = $6,
			game = $7, format = $8, rules = $9, visibility = $10, capacity = $11,
			start_at = $12, end_at = $13, timezone = $14, tags = $15, entry_fee = $16,
			language = $17, is_recurring = $18, recurrence_rule = $19, updated_at = $20
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		event.ID, event.HostUserID, event.GroupID, event.VenueID, event.Title, event.Description,
		event.Game, event.Format, rulesJSON, event.Visibility, event.Capacity,
		event.StartAt, event.EndAt, event.Timezone, event.Tags, event.EntryFee,
		event.Language, event.IsRecurring, event.RecurrenceRule, event.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

// Delete deletes an event
func (r *eventRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM events WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("event not found")
	}

	return nil
}

// Search searches for events based on parameters
func (r *eventRepository) Search(ctx context.Context, params domain.EventSearchParams) ([]*domain.Event, error) {
	query, args := r.buildSearchQuery(params, false)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// SearchWithDetails searches for events with details based on parameters
func (r *eventRepository) SearchWithDetails(ctx context.Context, params domain.EventSearchParams) ([]*domain.EventWithDetails, error) {
	// For now, implement as basic search and then fetch details for each
	events, err := r.Search(ctx, params)
	if err != nil {
		return nil, err
	}

	var eventsWithDetails []*domain.EventWithDetails
	for _, event := range events {
		eventWithDetails, err := r.GetByIDWithDetails(ctx, event.ID)
		if err != nil {
			return nil, err
		}
		eventsWithDetails = append(eventsWithDetails, eventWithDetails)
	}

	return eventsWithDetails, nil
}

// SearchNearby searches for events near a location using PostGIS
func (r *eventRepository) SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.Event, error) {
	query, args := r.buildNearbySearchQuery(lat, lon, radiusKm, params, false)
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search nearby events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// SearchNearbyWithDetails searches for events near a location with details
func (r *eventRepository) SearchNearbyWithDetails(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.EventWithDetails, error) {
	events, err := r.SearchNearby(ctx, lat, lon, radiusKm, params)
	if err != nil {
		return nil, err
	}

	var eventsWithDetails []*domain.EventWithDetails
	for _, event := range events {
		eventWithDetails, err := r.GetByIDWithDetails(ctx, event.ID)
		if err != nil {
			return nil, err
		}
		eventsWithDetails = append(eventsWithDetails, eventWithDetails)
	}

	return eventsWithDetails, nil
}

// GetUserEvents retrieves events for a specific user
func (r *eventRepository) GetUserEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Event, error) {
	query := `
		SELECT id, host_user_id, group_id, venue_id, title, description, game, format,
			rules, visibility, capacity, start_at, end_at, timezone, tags, entry_fee, language,
			is_recurring, recurrence_rule, created_at, updated_at
		FROM events
		WHERE host_user_id = $1
		ORDER BY start_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get user events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// GetGroupEvents retrieves events for a specific group
func (r *eventRepository) GetGroupEvents(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*domain.Event, error) {
	query := `
		SELECT id, host_user_id, group_id, venue_id, title, description, game, format,
			rules, visibility, capacity, start_at, end_at, timezone, tags, entry_fee, language,
			is_recurring, recurrence_rule, created_at, updated_at
		FROM events
		WHERE group_id = $1
		ORDER BY start_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get group events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// GetUpcomingEvents retrieves upcoming events
func (r *eventRepository) GetUpcomingEvents(ctx context.Context, limit, offset int) ([]*domain.Event, error) {
	query := `
		SELECT id, host_user_id, group_id, venue_id, title, description, game, format,
			rules, visibility, capacity, start_at, end_at, timezone, tags, entry_fee, language,
			is_recurring, recurrence_rule, created_at, updated_at
		FROM events
		WHERE start_at > NOW() AND visibility = 'public'
		ORDER BY start_at ASC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get upcoming events: %w", err)
	}
	defer rows.Close()

	var events []*domain.Event
	for rows.Next() {
		event, err := r.scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	return events, nil
}

// CreateRSVP creates a new event RSVP
func (r *eventRepository) CreateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error {
	query := `
		INSERT INTO event_rsvp (event_id, user_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`

	_, err := r.db.Exec(ctx, query,
		rsvp.EventID,
		rsvp.UserID,
		rsvp.Status,
		rsvp.CreatedAt,
		rsvp.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create RSVP: %w", err)
	}

	return nil
}

// GetRSVP retrieves an RSVP for a specific event and user
func (r *eventRepository) GetRSVP(ctx context.Context, eventID, userID uuid.UUID) (*domain.EventRSVP, error) {
	query := `
		SELECT event_id, user_id, status, created_at, updated_at
		FROM event_rsvp
		WHERE event_id = $1 AND user_id = $2`

	var rsvp domain.EventRSVP
	err := r.db.QueryRow(ctx, query, eventID, userID).Scan(
		&rsvp.EventID,
		&rsvp.UserID,
		&rsvp.Status,
		&rsvp.CreatedAt,
		&rsvp.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get RSVP: %w", err)
	}

	return &rsvp, nil
}

// UpdateRSVP updates an event RSVP
func (r *eventRepository) UpdateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error {
	query := `
		UPDATE event_rsvp
		SET status = $3, updated_at = $4
		WHERE event_id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query,
		rsvp.EventID,
		rsvp.UserID,
		rsvp.Status,
		rsvp.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update RSVP: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("RSVP not found")
	}

	return nil
}

// DeleteRSVP deletes an event RSVP
func (r *eventRepository) DeleteRSVP(ctx context.Context, eventID, userID uuid.UUID) error {
	query := `DELETE FROM event_rsvp WHERE event_id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, eventID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete RSVP: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("RSVP not found")
	}

	return nil
}

// GetEventRSVPs retrieves all RSVPs for an event
func (r *eventRepository) GetEventRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error) {
	query := `
		SELECT event_id, user_id, status, created_at, updated_at
		FROM event_rsvp
		WHERE event_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get event RSVPs: %w", err)
	}
	defer rows.Close()

	var rsvps []*domain.EventRSVP
	for rows.Next() {
		var rsvp domain.EventRSVP
		if err := rows.Scan(&rsvp.EventID, &rsvp.UserID, &rsvp.Status, &rsvp.CreatedAt, &rsvp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan RSVP: %w", err)
		}
		rsvps = append(rsvps, &rsvp)
	}

	return rsvps, nil
}

// GetUserRSVPs retrieves all RSVPs for a user
func (r *eventRepository) GetUserRSVPs(ctx context.Context, userID uuid.UUID) ([]*domain.EventRSVP, error) {
	query := `
		SELECT event_id, user_id, status, created_at, updated_at
		FROM event_rsvp
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user RSVPs: %w", err)
	}
	defer rows.Close()

	var rsvps []*domain.EventRSVP
	for rows.Next() {
		var rsvp domain.EventRSVP
		if err := rows.Scan(&rsvp.EventID, &rsvp.UserID, &rsvp.Status, &rsvp.CreatedAt, &rsvp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan RSVP: %w", err)
		}
		rsvps = append(rsvps, &rsvp)
	}

	return rsvps, nil
}

// CountRSVPsByStatus counts RSVPs by status for an event
func (r *eventRepository) CountRSVPsByStatus(ctx context.Context, eventID uuid.UUID, status domain.RSVPStatus) (int, error) {
	query := `SELECT COUNT(*) FROM event_rsvp WHERE event_id = $1 AND status = $2`

	var count int
	err := r.db.QueryRow(ctx, query, eventID, status).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count RSVPs by status: %w", err)
	}

	return count, nil
}

// GetWaitlistedRSVPs retrieves waitlisted RSVPs for an event
func (r *eventRepository) GetWaitlistedRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error) {
	query := `
		SELECT event_id, user_id, status, created_at, updated_at
		FROM event_rsvp
		WHERE event_id = $1 AND status = 'waitlisted'
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, eventID)
	if err != nil {
		return nil, fmt.Errorf("failed to get waitlisted RSVPs: %w", err)
	}
	defer rows.Close()

	var rsvps []*domain.EventRSVP
	for rows.Next() {
		var rsvp domain.EventRSVP
		if err := rows.Scan(&rsvp.EventID, &rsvp.UserID, &rsvp.Status, &rsvp.CreatedAt, &rsvp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan RSVP: %w", err)
		}
		rsvps = append(rsvps, &rsvp)
	}

	return rsvps, nil
}

// GetEventAttendeeCount gets the total number of attendees (going + waitlisted)
func (r *eventRepository) GetEventAttendeeCount(ctx context.Context, eventID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM event_rsvp WHERE event_id = $1 AND status IN ('going', 'waitlisted')`

	var count int
	err := r.db.QueryRow(ctx, query, eventID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get event attendee count: %w", err)
	}

	return count, nil
}

// GetEventGoingCount gets the number of people going to an event
func (r *eventRepository) GetEventGoingCount(ctx context.Context, eventID uuid.UUID) (int, error) {
	return r.CountRSVPsByStatus(ctx, eventID, domain.RSVPStatusGoing)
}

// Helper function to scan an event from a row
func (r *eventRepository) scanEvent(row pgx.Row) (*domain.Event, error) {
	var event domain.Event
	var rulesJSON []byte

	err := row.Scan(
		&event.ID,
		&event.HostUserID,
		&event.GroupID,
		&event.VenueID,
		&event.Title,
		&event.Description,
		&event.Game,
		&event.Format,
		&rulesJSON,
		&event.Visibility,
		&event.Capacity,
		&event.StartAt,
		&event.EndAt,
		&event.Timezone,
		&event.Tags,
		&event.EntryFee,
		&event.Language,
		&event.IsRecurring,
		&event.RecurrenceRule,
		&event.CreatedAt,
		&event.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to scan event: %w", err)
	}

	// Unmarshal rules
	if err := json.Unmarshal(rulesJSON, &event.Rules); err != nil {
		return nil, fmt.Errorf("failed to unmarshal rules: %w", err)
	}

	return &event, nil
}

// Helper function to build search query
func (r *eventRepository) buildSearchQuery(params domain.EventSearchParams, withDetails bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT id, host_user_id, group_id, venue_id, title, description, game, format,
			rules, visibility, capacity, start_at, end_at, timezone, tags, entry_fee, language,
			is_recurring, recurrence_rule, created_at, updated_at
		FROM events`

	// Add WHERE conditions
	if params.StartFrom != nil {
		conditions = append(conditions, fmt.Sprintf("start_at >= $%d", argIndex))
		args = append(args, *params.StartFrom)
		argIndex++
	}

	if params.Days != nil && params.StartFrom != nil {
		endDate := params.StartFrom.AddDate(0, 0, *params.Days)
		conditions = append(conditions, fmt.Sprintf("start_at <= $%d", argIndex))
		args = append(args, endDate)
		argIndex++
	}

	if params.Game != nil {
		conditions = append(conditions, fmt.Sprintf("game = $%d", argIndex))
		args = append(args, *params.Game)
		argIndex++
	}

	if params.Format != nil {
		conditions = append(conditions, fmt.Sprintf("format = $%d", argIndex))
		args = append(args, *params.Format)
		argIndex++
	}

	if params.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("visibility = $%d", argIndex))
		args = append(args, *params.Visibility)
		argIndex++
	}

	if params.GroupID != nil {
		conditions = append(conditions, fmt.Sprintf("group_id = $%d", argIndex))
		args = append(args, *params.GroupID)
		argIndex++
	}

	// Build final query
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	baseQuery += " ORDER BY start_at ASC"

	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	return baseQuery, args
}

// Helper function to build nearby search query using PostGIS
func (r *eventRepository) buildNearbySearchQuery(lat, lon float64, radiusKm int, params domain.EventSearchParams, withDetails bool) (string, []interface{}) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	baseQuery := `
		SELECT e.id, e.host_user_id, e.group_id, e.venue_id, e.title, e.description, e.game, e.format,
			e.rules, e.visibility, e.capacity, e.start_at, e.end_at, e.timezone, e.tags, e.entry_fee, e.language,
			e.is_recurring, e.recurrence_rule, e.created_at, e.updated_at
		FROM events e
		WHERE e.location IS NOT NULL 
		AND ST_DWithin(e.location, ST_Point($1, $2)::geography, $3)`

	args = append(args, lon, lat, radiusKm*1000) // Convert km to meters
	argIndex = 4

	// Add additional conditions
	if params.StartFrom != nil {
		conditions = append(conditions, fmt.Sprintf("e.start_at >= $%d", argIndex))
		args = append(args, *params.StartFrom)
		argIndex++
	}

	if params.Days != nil && params.StartFrom != nil {
		endDate := params.StartFrom.AddDate(0, 0, *params.Days)
		conditions = append(conditions, fmt.Sprintf("e.start_at <= $%d", argIndex))
		args = append(args, endDate)
		argIndex++
	}

	if params.Game != nil {
		conditions = append(conditions, fmt.Sprintf("e.game = $%d", argIndex))
		args = append(args, *params.Game)
		argIndex++
	}

	if params.Format != nil {
		conditions = append(conditions, fmt.Sprintf("e.format = $%d", argIndex))
		args = append(args, *params.Format)
		argIndex++
	}

	if params.Visibility != nil {
		conditions = append(conditions, fmt.Sprintf("e.visibility = $%d", argIndex))
		args = append(args, *params.Visibility)
		argIndex++
	}

	if params.GroupID != nil {
		conditions = append(conditions, fmt.Sprintf("e.group_id = $%d", argIndex))
		args = append(args, *params.GroupID)
		argIndex++
	}

	// Add additional conditions
	if len(conditions) > 0 {
		baseQuery += " AND " + strings.Join(conditions, " AND ")
	}

	// Order by distance and start time
	baseQuery += " ORDER BY ST_Distance(e.location, ST_Point($1, $2)::geography), e.start_at ASC"

	// Add pagination
	baseQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, params.Limit, params.Offset)

	return baseQuery, args
}
