package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
)

var (
	ErrUserNotFound    = errors.New("user not found")
	ErrProfileNotFound = errors.New("profile not found")
	ErrUnauthorized    = errors.New("unauthorized access")
)

// UpdateProfileRequest represents the request to update user profile
type UpdateProfileRequest struct {
	UserID                   uuid.UUID              `json:"-"` // Set from authentication context
	DisplayName              *string                `json:"display_name,omitempty" validate:"omitempty,max=100"`
	Locale                   *string                `json:"locale,omitempty" validate:"omitempty,oneof=en pt"`
	Timezone                 *string                `json:"timezone,omitempty"`
	Country                  *string                `json:"country,omitempty"`
	City                     *string                `json:"city,omitempty"`
	PreferredGames           []string               `json:"preferred_games,omitempty"`
	CommunicationPreferences map[string]interface{} `json:"communication_preferences,omitempty"`
	VisibilitySettings       map[string]interface{} `json:"visibility_settings,omitempty"`
}

// GetUserProfileRequest represents the request to get user profile
type GetUserProfileRequest struct {
	UserID         uuid.UUID `json:"-"` // User requesting the profile
	TargetUserID   uuid.UUID `json:"-"` // User whose profile is being requested
	IncludePrivate bool      `json:"-"` // Whether to include private information
}

// UserProfileResponse represents the user profile response
type UserProfileResponse struct {
	User    *domain.User    `json:"user"`
	Profile *domain.Profile `json:"profile"`
}

// UpdateProfileUseCase handles profile updates
type UpdateProfileUseCase struct {
	userRepo repository.UserRepository
}

// NewUpdateProfileUseCase creates a new UpdateProfileUseCase
func NewUpdateProfileUseCase(userRepo repository.UserRepository) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		userRepo: userRepo,
	}
}

// Execute updates the user's profile with the provided information
func (uc *UpdateProfileUseCase) Execute(ctx context.Context, req *UpdateProfileRequest) (*UserProfileResponse, error) {
	// Get existing profile
	existingProfile, err := uc.userRepo.GetProfile(ctx, req.UserID)
	if err != nil {
		return nil, ErrProfileNotFound
	}

	// Create updated profile with existing values as defaults
	updatedProfile := &domain.Profile{
		UserID:                   existingProfile.UserID,
		DisplayName:              existingProfile.DisplayName,
		Locale:                   existingProfile.Locale,
		Timezone:                 existingProfile.Timezone,
		Country:                  existingProfile.Country,
		City:                     existingProfile.City,
		PreferredGames:           existingProfile.PreferredGames,
		CommunicationPreferences: existingProfile.CommunicationPreferences,
		VisibilitySettings:       existingProfile.VisibilitySettings,
		UpdatedAt:                time.Now().UTC(),
	}

	// Update only provided fields
	if req.DisplayName != nil {
		updatedProfile.DisplayName = req.DisplayName
	}
	if req.Locale != nil {
		updatedProfile.Locale = *req.Locale
	}
	if req.Timezone != nil {
		updatedProfile.Timezone = *req.Timezone
	}
	if req.Country != nil {
		updatedProfile.Country = req.Country
	}
	if req.City != nil {
		updatedProfile.City = req.City
	}
	if req.PreferredGames != nil {
		updatedProfile.PreferredGames = req.PreferredGames
	}
	if req.CommunicationPreferences != nil {
		updatedProfile.CommunicationPreferences = req.CommunicationPreferences
	}
	if req.VisibilitySettings != nil {
		updatedProfile.VisibilitySettings = req.VisibilitySettings
	}

	// Validate updated profile
	if err := updatedProfile.Validate(); err != nil {
		return nil, err
	}

	// Update profile in repository
	if err := uc.userRepo.UpdateProfile(ctx, updatedProfile); err != nil {
		return nil, err
	}

	// Get user information
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return &UserProfileResponse{
		User:    user,
		Profile: updatedProfile,
	}, nil
}

// GetUserProfileUseCase handles retrieving user profiles
type GetUserProfileUseCase struct {
	userRepo repository.UserRepository
}

// NewGetUserProfileUseCase creates a new GetUserProfileUseCase
func NewGetUserProfileUseCase(userRepo repository.UserRepository) *GetUserProfileUseCase {
	return &GetUserProfileUseCase{
		userRepo: userRepo,
	}
}

// Execute retrieves a user profile with appropriate privacy controls
func (uc *GetUserProfileUseCase) Execute(ctx context.Context, req *GetUserProfileRequest) (*UserProfileResponse, error) {
	// Get user with profile
	userWithProfile, err := uc.userRepo.GetUserWithProfile(ctx, req.TargetUserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	if userWithProfile.Profile == nil {
		return nil, ErrProfileNotFound
	}

	// Check if user is requesting their own profile
	isOwnProfile := req.UserID == req.TargetUserID

	// Apply privacy controls if not own profile
	if !isOwnProfile && !req.IncludePrivate {
		userWithProfile = uc.applyPrivacyControls(userWithProfile)
	}

	return &UserProfileResponse{
		User:    &userWithProfile.User,
		Profile: userWithProfile.Profile,
	}, nil
}

// applyPrivacyControls filters profile information based on visibility settings
func (uc *GetUserProfileUseCase) applyPrivacyControls(userWithProfile *domain.UserWithProfile) *domain.UserWithProfile {
	profile := userWithProfile.Profile
	visibilitySettings := profile.VisibilitySettings

	// Create a copy to avoid modifying the original
	filteredProfile := &domain.Profile{
		UserID:                   profile.UserID,
		DisplayName:              profile.DisplayName,
		Locale:                   profile.Locale,
		Timezone:                 profile.Timezone,
		Country:                  profile.Country,
		City:                     profile.City,
		PreferredGames:           profile.PreferredGames,
		CommunicationPreferences: make(map[string]interface{}),
		VisibilitySettings:       make(map[string]interface{}),
		UpdatedAt:                profile.UpdatedAt,
	}

	// Apply visibility controls
	if showEmail, ok := visibilitySettings["show_email"].(bool); !ok || !showEmail {
		// Don't include email in public view (it's already excluded from User JSON)
	}

	if showRealName, ok := visibilitySettings["show_real_name"].(bool); !ok || !showRealName {
		filteredProfile.DisplayName = nil
	}

	if showLocation, ok := visibilitySettings["show_location"].(bool); !ok || !showLocation {
		filteredProfile.Country = nil
		filteredProfile.City = nil
	}

	if showPreferredGames, ok := visibilitySettings["show_preferred_games"].(bool); !ok || !showPreferredGames {
		filteredProfile.PreferredGames = []string{}
	}

	// Never expose communication preferences or full visibility settings to other users
	filteredProfile.CommunicationPreferences = make(map[string]interface{})
	filteredProfile.VisibilitySettings = make(map[string]interface{})

	return &domain.UserWithProfile{
		User:    userWithProfile.User,
		Profile: filteredProfile,
	}
}
