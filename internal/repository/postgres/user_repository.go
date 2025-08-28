package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
)

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *pgxpool.Pool) repository.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, password_hash, created_at, updated_at, is_active, last_login)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
		user.IsActive,
		user.LastLogin,
	)

	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at, is_active, last_login
		FROM users
		WHERE id = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.LastLogin,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, password_hash, created_at, updated_at, is_active, last_login
		FROM users
		WHERE email = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.IsActive,
		&user.LastLogin,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $2, password_hash = $3, updated_at = $4, is_active = $5, last_login = $6
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.UpdatedAt,
		user.IsActive,
		user.LastLogin,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// Delete deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// CreateProfile creates a user profile
func (r *userRepository) CreateProfile(ctx context.Context, profile *domain.Profile) error {
	preferredGamesJSON, err := json.Marshal(profile.PreferredGames)
	if err != nil {
		return fmt.Errorf("failed to marshal preferred games: %w", err)
	}

	commPrefsJSON, err := json.Marshal(profile.CommunicationPreferences)
	if err != nil {
		return fmt.Errorf("failed to marshal communication preferences: %w", err)
	}

	visibilityJSON, err := json.Marshal(profile.VisibilitySettings)
	if err != nil {
		return fmt.Errorf("failed to marshal visibility settings: %w", err)
	}

	query := `
		INSERT INTO profiles (user_id, display_name, locale, timezone, country, city, 
			preferred_games, communication_preferences, visibility_settings, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = r.db.Exec(ctx, query,
		profile.UserID,
		profile.DisplayName,
		profile.Locale,
		profile.Timezone,
		profile.Country,
		profile.City,
		preferredGamesJSON,
		commPrefsJSON,
		visibilityJSON,
		profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create profile: %w", err)
	}

	return nil
}

// GetProfile retrieves a user profile
func (r *userRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*domain.Profile, error) {
	query := `
		SELECT user_id, display_name, locale, timezone, country, city,
			preferred_games, communication_preferences, visibility_settings, updated_at
		FROM profiles
		WHERE user_id = $1`

	var profile domain.Profile
	var preferredGamesJSON, commPrefsJSON, visibilityJSON []byte

	err := r.db.QueryRow(ctx, query, userID).Scan(
		&profile.UserID,
		&profile.DisplayName,
		&profile.Locale,
		&profile.Timezone,
		&profile.Country,
		&profile.City,
		&preferredGamesJSON,
		&commPrefsJSON,
		&visibilityJSON,
		&profile.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	// Unmarshal JSON fields
	if err := json.Unmarshal(preferredGamesJSON, &profile.PreferredGames); err != nil {
		return nil, fmt.Errorf("failed to unmarshal preferred games: %w", err)
	}

	if err := json.Unmarshal(commPrefsJSON, &profile.CommunicationPreferences); err != nil {
		return nil, fmt.Errorf("failed to unmarshal communication preferences: %w", err)
	}

	if err := json.Unmarshal(visibilityJSON, &profile.VisibilitySettings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal visibility settings: %w", err)
	}

	return &profile, nil
}

// UpdateProfile updates a user profile
func (r *userRepository) UpdateProfile(ctx context.Context, profile *domain.Profile) error {
	preferredGamesJSON, err := json.Marshal(profile.PreferredGames)
	if err != nil {
		return fmt.Errorf("failed to marshal preferred games: %w", err)
	}

	commPrefsJSON, err := json.Marshal(profile.CommunicationPreferences)
	if err != nil {
		return fmt.Errorf("failed to marshal communication preferences: %w", err)
	}

	visibilityJSON, err := json.Marshal(profile.VisibilitySettings)
	if err != nil {
		return fmt.Errorf("failed to marshal visibility settings: %w", err)
	}

	query := `
		UPDATE profiles
		SET display_name = $2, locale = $3, timezone = $4, country = $5, city = $6,
			preferred_games = $7, communication_preferences = $8, visibility_settings = $9, updated_at = $10
		WHERE user_id = $1`

	result, err := r.db.Exec(ctx, query,
		profile.UserID,
		profile.DisplayName,
		profile.Locale,
		profile.Timezone,
		profile.Country,
		profile.City,
		preferredGamesJSON,
		commPrefsJSON,
		visibilityJSON,
		profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update profile: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("profile not found")
	}

	return nil
}

// GetUserWithProfile retrieves a user with their profile
func (r *userRepository) GetUserWithProfile(ctx context.Context, userID uuid.UUID) (*domain.UserWithProfile, error) {
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}

	profile, err := r.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}

	return &domain.UserWithProfile{
		User:    *user,
		Profile: profile,
	}, nil
}

// ExportUserData exports all user data for GDPR compliance
func (r *userRepository) ExportUserData(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	data := make(map[string]interface{})

	// Get user data
	user, err := r.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user != nil {
		data["user"] = user
	}

	// Get profile data
	profile, err := r.GetProfile(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile != nil {
		data["profile"] = profile
	}

	// Get group memberships
	var groupMemberships []domain.GroupMember
	groupQuery := `SELECT group_id, user_id, role, joined_at FROM group_members WHERE user_id = $1`
	rows, err := tx.Query(ctx, groupQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group memberships: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var gm domain.GroupMember
		if err := rows.Scan(&gm.GroupID, &gm.UserID, &gm.Role, &gm.JoinedAt); err != nil {
			return nil, fmt.Errorf("failed to scan group membership: %w", err)
		}
		groupMemberships = append(groupMemberships, gm)
	}
	data["group_memberships"] = groupMemberships

	// Get event RSVPs
	var rsvps []domain.EventRSVP
	rsvpQuery := `SELECT event_id, user_id, status, created_at, updated_at FROM event_rsvp WHERE user_id = $1`
	rows, err = tx.Query(ctx, rsvpQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get RSVPs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rsvp domain.EventRSVP
		if err := rows.Scan(&rsvp.EventID, &rsvp.UserID, &rsvp.Status, &rsvp.CreatedAt, &rsvp.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan RSVP: %w", err)
		}
		rsvps = append(rsvps, rsvp)
	}
	data["event_rsvps"] = rsvps

	return data, nil
}

// DeleteUserData deletes all user data for GDPR compliance
func (r *userRepository) DeleteUserData(ctx context.Context, userID uuid.UUID) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Delete in order to respect foreign key constraints
	// RSVPs will be deleted by CASCADE
	// Group memberships will be deleted by CASCADE
	// Profile will be deleted by CASCADE
	// Events hosted by user will have host_user_id set to NULL or be deleted
	// Groups owned by user will be transferred or deleted
	// Venues created by user will have created_by set to NULL

	// Update venues to remove reference
	_, err = tx.Exec(ctx, "UPDATE venues SET created_by = NULL WHERE created_by = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to update venues: %w", err)
	}

	// Delete user (CASCADE will handle related data)
	_, err = tx.Exec(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return tx.Commit(ctx)
}

// UpdateLastLogin updates the user's last login time
func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, loginTime time.Time) error {
	query := `UPDATE users SET last_login = $2, updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, userID, loginTime)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// SetActive sets the user's active status
func (r *userRepository) SetActive(ctx context.Context, userID uuid.UUID, active bool) error {
	query := `UPDATE users SET is_active = $2, updated_at = NOW() WHERE id = $1`

	result, err := r.db.Exec(ctx, query, userID, active)
	if err != nil {
		return fmt.Errorf("failed to set user active status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
