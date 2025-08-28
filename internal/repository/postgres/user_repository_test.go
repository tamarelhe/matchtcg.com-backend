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

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Verify user was created
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, user.Email, retrieved.Email)
	assert.Equal(t, user.PasswordHash, retrieved.PasswordHash)
	assert.Equal(t, user.IsActive, retrieved.IsActive)
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Test GetByEmail
	retrieved, err := repo.GetByEmail(ctx, user.Email)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.ID)
	assert.Equal(t, user.Email, retrieved.Email)

	// Test non-existent email
	notFound, err := repo.GetByEmail(ctx, "nonexistent@example.com")
	require.NoError(t, err)
	assert.Nil(t, notFound)
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update user
	user.Email = "updated@example.com"
	user.IsActive = false
	user.UpdatedAt = time.Now()

	err = repo.Update(ctx, user)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, "updated@example.com", retrieved.Email)
	assert.False(t, retrieved.IsActive)
}

func TestUserRepository_Profile(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create user first
	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile
	displayName := "Test User"
	profile := &domain.Profile{
		UserID:                   user.ID,
		DisplayName:              &displayName,
		Locale:                   "en",
		Timezone:                 "UTC",
		PreferredGames:           []string{"mtg", "lorcana"},
		CommunicationPreferences: map[string]interface{}{"email": true},
		VisibilitySettings:       map[string]interface{}{"public": true},
		UpdatedAt:                time.Now(),
	}

	err = repo.CreateProfile(ctx, profile)
	require.NoError(t, err)

	// Get profile
	retrieved, err := repo.GetProfile(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, user.ID, retrieved.UserID)
	assert.Equal(t, "Test User", *retrieved.DisplayName)
	assert.Equal(t, "en", retrieved.Locale)
	assert.Equal(t, []string{"mtg", "lorcana"}, retrieved.PreferredGames)

	// Update profile
	newDisplayName := "Updated User"
	profile.DisplayName = &newDisplayName
	profile.Locale = "pt"
	profile.UpdatedAt = time.Now()

	err = repo.UpdateProfile(ctx, profile)
	require.NoError(t, err)

	// Verify update
	updated, err := repo.GetProfile(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, "Updated User", *updated.DisplayName)
	assert.Equal(t, "pt", updated.Locale)
}

func TestUserRepository_GetUserWithProfile(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	// Create user
	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Create profile
	displayName := "Test User"
	profile := &domain.Profile{
		UserID:                   user.ID,
		DisplayName:              &displayName,
		Locale:                   "en",
		Timezone:                 "UTC",
		PreferredGames:           []string{"mtg"},
		CommunicationPreferences: map[string]interface{}{"email": true},
		VisibilitySettings:       map[string]interface{}{"public": true},
		UpdatedAt:                time.Now(),
	}

	err = repo.CreateProfile(ctx, profile)
	require.NoError(t, err)

	// Get user with profile
	userWithProfile, err := repo.GetUserWithProfile(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, userWithProfile)
	assert.Equal(t, user.ID, userWithProfile.User.ID)
	assert.Equal(t, user.Email, userWithProfile.User.Email)
	require.NotNil(t, userWithProfile.Profile)
	assert.Equal(t, "Test User", *userWithProfile.Profile.DisplayName)
}

func TestUserRepository_UpdateLastLogin(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Update last login
	loginTime := time.Now()
	err = repo.UpdateLastLogin(ctx, user.ID, loginTime)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.NotNil(t, retrieved.LastLogin)
	assert.WithinDuration(t, loginTime, *retrieved.LastLogin, time.Second)
}

func TestUserRepository_SetActive(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Set inactive
	err = repo.SetActive(ctx, user.ID, false)
	require.NoError(t, err)

	// Verify update
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.False(t, retrieved.IsActive)
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewUserRepository(db)
	ctx := context.Background()

	user := &domain.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: "hashedpassword",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		IsActive:     true,
	}

	err := repo.Create(ctx, user)
	require.NoError(t, err)

	// Delete user
	err = repo.Delete(ctx, user.ID)
	require.NoError(t, err)

	// Verify deletion
	retrieved, err := repo.GetByID(ctx, user.ID)
	require.NoError(t, err)
	assert.Nil(t, retrieved)
}
