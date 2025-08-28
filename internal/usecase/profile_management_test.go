package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUpdateProfileUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	useCase := NewUpdateProfileUseCase(mockUserRepo)

	userID := uuid.New()
	displayName := "Updated Name"
	locale := "en"
	timezone := "Europe/Lisbon"
	country := "Portugal"
	city := "Porto"

	existingProfile := &domain.Profile{
		UserID:         userID,
		DisplayName:    stringPtr("Old Name"),
		Locale:         "pt",
		Timezone:       "UTC",
		Country:        stringPtr("Spain"),
		City:           stringPtr("Madrid"),
		PreferredGames: []string{"mtg"},
		CommunicationPreferences: map[string]interface{}{
			"email_notifications": true,
		},
		VisibilitySettings: map[string]interface{}{
			"profile_visibility": "public",
		},
		UpdatedAt: time.Now().UTC().Add(-time.Hour),
	}

	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
	}

	req := &UpdateProfileRequest{
		UserID:         userID,
		DisplayName:    &displayName,
		Locale:         &locale,
		Timezone:       &timezone,
		Country:        &country,
		City:           &city,
		PreferredGames: []string{"mtg", "lorcana"},
	}

	// Mock expectations
	mockUserRepo.On("GetProfile", mock.Anything, userID).Return(existingProfile, nil)
	mockUserRepo.On("UpdateProfile", mock.Anything, mock.AnythingOfType("*domain.Profile")).Return(nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, displayName, *result.Profile.DisplayName)
	assert.Equal(t, locale, result.Profile.Locale)
	assert.Equal(t, timezone, result.Profile.Timezone)
	assert.Equal(t, country, *result.Profile.Country)
	assert.Equal(t, city, *result.Profile.City)
	assert.Equal(t, []string{"mtg", "lorcana"}, result.Profile.PreferredGames)

	mockUserRepo.AssertExpectations(t)
}

func TestUpdateProfileUseCase_Execute_ProfileNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	useCase := NewUpdateProfileUseCase(mockUserRepo)

	userID := uuid.New()
	req := &UpdateProfileRequest{
		UserID: userID,
	}

	// Mock expectations
	mockUserRepo.On("GetProfile", mock.Anything, userID).Return((*domain.Profile)(nil), errors.New("not found"))

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrProfileNotFound, err)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
}

func TestUpdateProfileUseCase_Execute_PartialUpdate(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	useCase := NewUpdateProfileUseCase(mockUserRepo)

	userID := uuid.New()
	newDisplayName := "New Name"

	existingProfile := &domain.Profile{
		UserID:         userID,
		DisplayName:    stringPtr("Old Name"),
		Locale:         "pt",
		Timezone:       "UTC",
		Country:        stringPtr("Portugal"),
		City:           stringPtr("Lisbon"),
		PreferredGames: []string{"mtg"},
		CommunicationPreferences: map[string]interface{}{
			"email_notifications": true,
		},
		VisibilitySettings: map[string]interface{}{
			"profile_visibility": "public",
		},
		UpdatedAt: time.Now().UTC().Add(-time.Hour),
	}

	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
	}

	req := &UpdateProfileRequest{
		UserID:      userID,
		DisplayName: &newDisplayName,
		// Only updating display name, other fields should remain unchanged
	}

	// Mock expectations
	mockUserRepo.On("GetProfile", mock.Anything, userID).Return(existingProfile, nil)
	mockUserRepo.On("UpdateProfile", mock.Anything, mock.AnythingOfType("*domain.Profile")).Return(nil)
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newDisplayName, *result.Profile.DisplayName)
	assert.Equal(t, "pt", result.Profile.Locale)         // Should remain unchanged
	assert.Equal(t, "UTC", result.Profile.Timezone)      // Should remain unchanged
	assert.Equal(t, "Portugal", *result.Profile.Country) // Should remain unchanged

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserProfileUseCase_Execute_OwnProfile(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	useCase := NewGetUserProfileUseCase(mockUserRepo)

	userID := uuid.New()
	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
	}

	profile := &domain.Profile{
		UserID:         userID,
		DisplayName:    stringPtr("Test User"),
		Locale:         "en",
		Timezone:       "UTC",
		Country:        stringPtr("Portugal"),
		City:           stringPtr("Lisbon"),
		PreferredGames: []string{"mtg"},
		CommunicationPreferences: map[string]interface{}{
			"email_notifications": true,
		},
		VisibilitySettings: map[string]interface{}{
			"show_email":     false,
			"show_real_name": true,
		},
		UpdatedAt: time.Now().UTC(),
	}

	userWithProfile := &domain.UserWithProfile{
		User:    *user,
		Profile: profile,
	}

	req := &GetUserProfileRequest{
		UserID:       userID,
		TargetUserID: userID, // Same user - own profile
	}

	// Mock expectations
	mockUserRepo.On("GetUserWithProfile", mock.Anything, userID).Return(userWithProfile, nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.User.ID)
	assert.Equal(t, "Test User", *result.Profile.DisplayName)
	// Should include all information for own profile
	assert.NotEmpty(t, result.Profile.CommunicationPreferences)
	assert.NotEmpty(t, result.Profile.VisibilitySettings)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserProfileUseCase_Execute_OtherUserProfile_PublicView(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	useCase := NewGetUserProfileUseCase(mockUserRepo)

	requestingUserID := uuid.New()
	targetUserID := uuid.New()

	user := &domain.User{
		ID:    targetUserID,
		Email: "target@example.com",
	}

	profile := &domain.Profile{
		UserID:         targetUserID,
		DisplayName:    stringPtr("Target User"),
		Locale:         "en",
		Timezone:       "UTC",
		Country:        stringPtr("Portugal"),
		City:           stringPtr("Lisbon"),
		PreferredGames: []string{"mtg"},
		CommunicationPreferences: map[string]interface{}{
			"email_notifications": true,
		},
		VisibilitySettings: map[string]interface{}{
			"show_email":           false,
			"show_real_name":       true,
			"show_location":        true,
			"show_preferred_games": true,
		},
		UpdatedAt: time.Now().UTC(),
	}

	userWithProfile := &domain.UserWithProfile{
		User:    *user,
		Profile: profile,
	}

	req := &GetUserProfileRequest{
		UserID:         requestingUserID,
		TargetUserID:   targetUserID,
		IncludePrivate: false, // Public view
	}

	// Mock expectations
	mockUserRepo.On("GetUserWithProfile", mock.Anything, targetUserID).Return(userWithProfile, nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, targetUserID, result.User.ID)
	assert.Equal(t, "Target User", *result.Profile.DisplayName)
	// Should not include private information
	assert.Empty(t, result.Profile.CommunicationPreferences)
	assert.Empty(t, result.Profile.VisibilitySettings)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserProfileUseCase_Execute_OtherUserProfile_HiddenName(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	useCase := NewGetUserProfileUseCase(mockUserRepo)

	requestingUserID := uuid.New()
	targetUserID := uuid.New()

	user := &domain.User{
		ID:    targetUserID,
		Email: "target@example.com",
	}

	profile := &domain.Profile{
		UserID:         targetUserID,
		DisplayName:    stringPtr("Target User"),
		Locale:         "en",
		Timezone:       "UTC",
		Country:        stringPtr("Portugal"),
		City:           stringPtr("Lisbon"),
		PreferredGames: []string{"mtg"},
		CommunicationPreferences: map[string]interface{}{
			"email_notifications": true,
		},
		VisibilitySettings: map[string]interface{}{
			"show_email":           false,
			"show_real_name":       false, // Hidden
			"show_location":        false, // Hidden
			"show_preferred_games": false, // Hidden
		},
		UpdatedAt: time.Now().UTC(),
	}

	userWithProfile := &domain.UserWithProfile{
		User:    *user,
		Profile: profile,
	}

	req := &GetUserProfileRequest{
		UserID:         requestingUserID,
		TargetUserID:   targetUserID,
		IncludePrivate: false,
	}

	// Mock expectations
	mockUserRepo.On("GetUserWithProfile", mock.Anything, targetUserID).Return(userWithProfile, nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, targetUserID, result.User.ID)
	// Should hide information based on visibility settings
	assert.Nil(t, result.Profile.DisplayName)
	assert.Nil(t, result.Profile.Country)
	assert.Nil(t, result.Profile.City)
	assert.Empty(t, result.Profile.PreferredGames)

	mockUserRepo.AssertExpectations(t)
}

func TestGetUserProfileUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	useCase := NewGetUserProfileUseCase(mockUserRepo)

	userID := uuid.New()
	targetUserID := uuid.New()

	req := &GetUserProfileRequest{
		UserID:       userID,
		TargetUserID: targetUserID,
	}

	// Mock expectations
	mockUserRepo.On("GetUserWithProfile", mock.Anything, targetUserID).Return((*domain.UserWithProfile)(nil), errors.New("not found"))

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
