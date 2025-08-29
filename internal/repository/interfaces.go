package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	// User operations
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Profile operations
	CreateProfile(ctx context.Context, profile *domain.Profile) error
	GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error)
	UpdateProfile(ctx context.Context, profile *domain.Profile) error
	GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*domain.UserWithProfile, error)
	CreateUserWithProfile(ctx context.Context, user *domain.User, profile *domain.Profile) error

	// GDPR compliance methods
	ExportUserData(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error)
	DeleteUserData(ctx context.Context, userID uuid.UUID) error

	// Authentication support
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error
	SetActive(ctx context.Context, userID uuid.UUID, active bool) error
}

// EventRepository defines the interface for event data operations
type EventRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, event *domain.Event) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Event, error)
	GetByIDWithDetails(ctx context.Context, id uuid.UUID) (*domain.EventWithDetails, error)
	Update(ctx context.Context, event *domain.Event) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Search operations
	Search(ctx context.Context, params domain.EventSearchParams) ([]*domain.Event, error)
	SearchWithDetails(ctx context.Context, params domain.EventSearchParams) ([]*domain.EventWithDetails, error)
	SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.Event, error)
	SearchNearbyWithDetails(ctx context.Context, lat, lon float64, radiusKm int, params domain.EventSearchParams) ([]*domain.EventWithDetails, error)

	// Event-specific queries
	GetUserEvents(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Event, error)
	GetGroupEvents(ctx context.Context, groupID uuid.UUID, limit, offset int) ([]*domain.Event, error)
	GetUpcomingEvents(ctx context.Context, limit, offset int) ([]*domain.Event, error)

	// RSVP operations
	CreateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error
	GetRSVP(ctx context.Context, eventID, userID uuid.UUID) (*domain.EventRSVP, error)
	UpdateRSVP(ctx context.Context, rsvp *domain.EventRSVP) error
	DeleteRSVP(ctx context.Context, eventID, userID uuid.UUID) error
	GetEventRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error)
	GetUserRSVPs(ctx context.Context, userID uuid.UUID) ([]*domain.EventRSVP, error)
	CountRSVPsByStatus(ctx context.Context, eventID uuid.UUID, status domain.RSVPStatus) (int, error)
	GetWaitlistedRSVPs(ctx context.Context, eventID uuid.UUID) ([]*domain.EventRSVP, error)

	// Capacity management
	GetEventAttendeeCount(ctx context.Context, eventID uuid.UUID) (int, error)
	GetEventGoingCount(ctx context.Context, eventID uuid.UUID) (int, error)
}

// GroupRepository defines the interface for group data operations
type GroupRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, group *domain.Group) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, error)
	GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*domain.GroupWithMembers, error)
	Update(ctx context.Context, group *domain.Group) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Group queries
	GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*domain.Group, error)
	GetUserGroupsWithMembers(ctx context.Context, userID uuid.UUID) ([]*domain.GroupWithMembers, error)
	GetGroupsByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Group, error)

	// Member management
	AddMember(ctx context.Context, member *domain.GroupMember) error
	GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error)
	UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role domain.GroupRole) error
	RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error
	GetGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMember, error)
	GetMembersByRole(ctx context.Context, groupID uuid.UUID, role domain.GroupRole) ([]*domain.GroupMember, error)
	IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (domain.GroupRole, error)

	// Permission checks
	CanUserAccessGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	CanUserManageGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error)
	IsGroupOwner(ctx context.Context, groupID, userID uuid.UUID) (bool, error)

	// Group statistics
	GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error)
}

// VenueRepository defines the interface for venue data operations
type VenueRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, venue *domain.Venue) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Venue, error)
	Update(ctx context.Context, venue *domain.Venue) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Location-based queries
	SearchNearby(ctx context.Context, lat, lon float64, radiusKm int, limit, offset int) ([]*domain.Venue, error)
	SearchByCity(ctx context.Context, city string, limit, offset int) ([]*domain.Venue, error)
	SearchByCountry(ctx context.Context, country string, limit, offset int) ([]*domain.Venue, error)
	SearchByName(ctx context.Context, name string, limit, offset int) ([]*domain.Venue, error)

	// Venue queries
	GetByCreator(ctx context.Context, creatorID uuid.UUID, limit, offset int) ([]*domain.Venue, error)
	GetByType(ctx context.Context, venueType domain.VenueType, limit, offset int) ([]*domain.Venue, error)
	GetPopularVenues(ctx context.Context, limit, offset int) ([]*domain.Venue, error)

	// Coordinate operations
	GetVenuesInBounds(ctx context.Context, northLat, southLat, eastLon, westLon float64) ([]*domain.Venue, error)
	FindNearestVenue(ctx context.Context, lat, lon float64) (*domain.Venue, error)

	// Statistics
	CountVenuesByCity(ctx context.Context, city string) (int, error)
	CountVenuesByCountry(ctx context.Context, country string) (int, error)
}

// NotificationRepository defines the interface for notification data operations
type NotificationRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, notification *domain.Notification) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error)
	Update(ctx context.Context, notification *domain.Notification) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Notification queries
	GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error)
	GetPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)
	GetFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error)

	// Status management
	MarkAsSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error
	MarkAsFailed(ctx context.Context, id uuid.UUID, errorMessage string) error
	IncrementRetryCount(ctx context.Context, id uuid.UUID) error

	// Cleanup operations
	DeleteOldNotifications(ctx context.Context, olderThan time.Time) error
}
