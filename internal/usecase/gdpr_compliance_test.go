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

// MockNotificationRepository is a mock implementation of NotificationRepository
type MockNotificationRepository struct {
	mock.Mock
}

func (m *MockNotificationRepository) Create(ctx context.Context, notification *domain.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Notification, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Notification), args.Error(1)
}

func (m *MockNotificationRepository) Update(ctx context.Context, notification *domain.Notification) error {
	args := m.Called(ctx, notification)
	return args.Error(0)
}

func (m *MockNotificationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepository) GetUserNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*domain.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetPendingNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.Notification), args.Error(1)
}

func (m *MockNotificationRepository) GetFailedNotifications(ctx context.Context, limit int) ([]*domain.Notification, error) {
	args := m.Called(ctx, limit)
	return args.Get(0).([]*domain.Notification), args.Error(1)
}

func (m *MockNotificationRepository) MarkAsSent(ctx context.Context, id uuid.UUID, sentAt time.Time) error {
	args := m.Called(ctx, id, sentAt)
	return args.Error(0)
}

func (m *MockNotificationRepository) MarkAsFailed(ctx context.Context, id uuid.UUID, errorMessage string) error {
	args := m.Called(ctx, id, errorMessage)
	return args.Error(0)
}

func (m *MockNotificationRepository) IncrementRetryCount(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockNotificationRepository) DeleteOldNotifications(ctx context.Context, olderThan time.Time) error {
	args := m.Called(ctx, olderThan)
	return args.Error(0)
}

func TestExportUserDataUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockEventRepo := new(MockEventRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockNotificationRepo := new(MockNotificationRepository)

	useCase := NewExportUserDataUseCase(mockUserRepo, mockEventRepo, mockGroupRepo, mockNotificationRepo)

	userID := uuid.New()
	user := &domain.User{
		ID:        userID,
		Email:     "test@example.com",
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
		IsActive:  true,
	}

	profile := &domain.Profile{
		UserID:      userID,
		DisplayName: stringPtr("Test User"),
		Locale:      "en",
		Timezone:    "UTC",
		UpdatedAt:   time.Now().UTC(),
	}

	userWithProfile := &domain.UserWithProfile{
		User:    *user,
		Profile: profile,
	}

	req := &ExportUserDataRequest{
		UserID: userID,
	}

	// Mock expectations
	mockUserRepo.On("GetUserWithProfile", mock.Anything, userID).Return(userWithProfile, nil)
	mockEventRepo.On("GetUserEvents", mock.Anything, userID, 1000, 0).Return([]*domain.Event{}, nil)
	mockEventRepo.On("GetUserRSVPs", mock.Anything, userID).Return([]*domain.EventRSVP{}, nil)
	mockGroupRepo.On("GetUserGroups", mock.Anything, userID).Return([]*domain.Group{}, nil)
	mockNotificationRepo.On("GetUserNotifications", mock.Anything, userID, 1000, 0).Return([]*domain.Notification{}, nil)
	mockUserRepo.On("ExportUserData", mock.Anything, userID).Return(map[string]interface{}{}, nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.NotNil(t, result.Data["account"])
	assert.NotNil(t, result.Data["profile"])
	assert.NotNil(t, result.Data["hosted_events"])
	assert.NotNil(t, result.Data["event_rsvps"])
	assert.NotNil(t, result.Data["groups"])
	assert.NotNil(t, result.Data["group_memberships"])
	assert.NotNil(t, result.Data["notifications"])

	mockUserRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
}

func TestExportUserDataUseCase_Execute_UserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockEventRepo := new(MockEventRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockNotificationRepo := new(MockNotificationRepository)

	useCase := NewExportUserDataUseCase(mockUserRepo, mockEventRepo, mockGroupRepo, mockNotificationRepo)

	userID := uuid.New()
	req := &ExportUserDataRequest{
		UserID: userID,
	}

	// Mock expectations
	mockUserRepo.On("GetUserWithProfile", mock.Anything, userID).Return((*domain.UserWithProfile)(nil), errors.New("not found"))

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, ErrUserNotFound, err)
	assert.Nil(t, result)

	mockUserRepo.AssertExpectations(t)
}

func TestDeleteUserAccountUseCase_Execute_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockEventRepo := new(MockEventRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockNotificationRepo := new(MockNotificationRepository)

	useCase := NewDeleteUserAccountUseCase(mockUserRepo, mockEventRepo, mockGroupRepo, mockNotificationRepo)

	userID := uuid.New()
	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
	}

	req := &DeleteUserAccountRequest{
		UserID: userID,
	}

	// Mock expectations
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockGroupRepo.On("GetGroupsByOwner", mock.Anything, userID).Return([]*domain.Group{}, nil)
	mockEventRepo.On("GetUserEvents", mock.Anything, userID, 1000, 0).Return([]*domain.Event{}, nil)
	mockGroupRepo.On("GetUserGroups", mock.Anything, userID).Return([]*domain.Group{}, nil)
	mockEventRepo.On("GetUserRSVPs", mock.Anything, userID).Return([]*domain.EventRSVP{}, nil)
	mockUserRepo.On("DeleteUserData", mock.Anything, userID).Return(nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Contains(t, result.Message, "permanently deleted")

	mockUserRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
}

func TestDeleteUserAccountUseCase_Execute_WithGroupOwnership(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	mockEventRepo := new(MockEventRepository)
	mockGroupRepo := new(MockGroupRepository)
	mockNotificationRepo := new(MockNotificationRepository)

	useCase := NewDeleteUserAccountUseCase(mockUserRepo, mockEventRepo, mockGroupRepo, mockNotificationRepo)

	userID := uuid.New()
	adminUserID := uuid.New()
	groupID := uuid.New()

	user := &domain.User{
		ID:    userID,
		Email: "test@example.com",
	}

	ownedGroup := &domain.Group{
		ID:          groupID,
		Name:        "Test Group",
		OwnerUserID: userID,
	}

	adminMember := &domain.GroupMember{
		GroupID: groupID,
		UserID:  adminUserID,
		Role:    domain.GroupRoleAdmin,
	}

	req := &DeleteUserAccountRequest{
		UserID: userID,
	}

	// Mock expectations
	mockUserRepo.On("GetByID", mock.Anything, userID).Return(user, nil)
	mockGroupRepo.On("GetGroupsByOwner", mock.Anything, userID).Return([]*domain.Group{ownedGroup}, nil)
	mockGroupRepo.On("GetGroupMembers", mock.Anything, groupID).Return([]*domain.GroupMember{adminMember}, nil)
	mockGroupRepo.On("Update", mock.Anything, mock.AnythingOfType("*domain.Group")).Return(nil)
	mockGroupRepo.On("UpdateMemberRole", mock.Anything, groupID, adminUserID, domain.GroupRoleOwner).Return(nil)
	mockEventRepo.On("GetUserEvents", mock.Anything, userID, 1000, 0).Return([]*domain.Event{}, nil)
	mockGroupRepo.On("GetUserGroups", mock.Anything, userID).Return([]*domain.Group{}, nil)
	mockEventRepo.On("GetUserRSVPs", mock.Anything, userID).Return([]*domain.EventRSVP{}, nil)
	mockUserRepo.On("DeleteUserData", mock.Anything, userID).Return(nil)

	// Act
	result, err := useCase.Execute(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)

	mockUserRepo.AssertExpectations(t)
	mockEventRepo.AssertExpectations(t)
	mockGroupRepo.AssertExpectations(t)
}

func TestConsentManagementService_UpdateConsent_Success(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	service := NewConsentManagementService(mockUserRepo)

	userID := uuid.New()
	profile := &domain.Profile{
		UserID:                   userID,
		CommunicationPreferences: make(map[string]interface{}),
		UpdatedAt:                time.Now().UTC().Add(-time.Hour),
	}

	req := &ConsentUpdateRequest{
		UserID:      userID,
		ConsentType: "marketing_emails",
		Granted:     true,
		Metadata: map[string]interface{}{
			"source": "registration",
		},
	}

	// Mock expectations
	mockUserRepo.On("GetProfile", mock.Anything, userID).Return(profile, nil)
	mockUserRepo.On("UpdateProfile", mock.Anything, mock.AnythingOfType("*domain.Profile")).Return(nil)

	// Act
	result, err := service.UpdateConsent(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "marketing_emails", result.ConsentType)
	assert.True(t, result.Granted)
	assert.Nil(t, result.RevokedAt)
	assert.NotNil(t, result.Metadata)

	mockUserRepo.AssertExpectations(t)
}

func TestConsentManagementService_UpdateConsent_Revoke(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	service := NewConsentManagementService(mockUserRepo)

	userID := uuid.New()
	profile := &domain.Profile{
		UserID:                   userID,
		CommunicationPreferences: make(map[string]interface{}),
		UpdatedAt:                time.Now().UTC().Add(-time.Hour),
	}

	req := &ConsentUpdateRequest{
		UserID:      userID,
		ConsentType: "marketing_emails",
		Granted:     false, // Revoking consent
	}

	// Mock expectations
	mockUserRepo.On("GetProfile", mock.Anything, userID).Return(profile, nil)
	mockUserRepo.On("UpdateProfile", mock.Anything, mock.AnythingOfType("*domain.Profile")).Return(nil)

	// Act
	result, err := service.UpdateConsent(context.Background(), req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "marketing_emails", result.ConsentType)
	assert.False(t, result.Granted)
	assert.NotNil(t, result.RevokedAt) // Should have revocation timestamp

	mockUserRepo.AssertExpectations(t)
}

func TestConsentManagementService_HasValidConsent_True(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	service := NewConsentManagementService(mockUserRepo)

	userID := uuid.New()

	// Create a profile with existing consent
	consentRecord := &ConsentRecord{
		ID:          uuid.New(),
		UserID:      userID,
		ConsentType: "marketing_emails",
		Granted:     true,
		GrantedAt:   time.Now().UTC(),
		RevokedAt:   nil,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	profile := &domain.Profile{
		UserID: userID,
		CommunicationPreferences: map[string]interface{}{
			"consent_marketing_emails": map[string]interface{}{
				"id":           consentRecord.ID.String(),
				"user_id":      consentRecord.UserID.String(),
				"consent_type": consentRecord.ConsentType,
				"granted":      consentRecord.Granted,
				"granted_at":   consentRecord.GrantedAt,
				"revoked_at":   consentRecord.RevokedAt,
				"created_at":   consentRecord.CreatedAt,
				"updated_at":   consentRecord.UpdatedAt,
			},
		},
		UpdatedAt: time.Now().UTC(),
	}

	// Mock expectations
	mockUserRepo.On("GetProfile", mock.Anything, userID).Return(profile, nil)

	// Act
	hasConsent, err := service.HasValidConsent(context.Background(), userID, "marketing_emails")

	// Assert
	assert.NoError(t, err)
	assert.True(t, hasConsent)

	mockUserRepo.AssertExpectations(t)
}

func TestConsentManagementService_HasValidConsent_False_NoConsent(t *testing.T) {
	// Arrange
	mockUserRepo := new(MockUserRepository)
	service := NewConsentManagementService(mockUserRepo)

	userID := uuid.New()
	profile := &domain.Profile{
		UserID:                   userID,
		CommunicationPreferences: make(map[string]interface{}),
		UpdatedAt:                time.Now().UTC(),
	}

	// Mock expectations
	mockUserRepo.On("GetProfile", mock.Anything, userID).Return(profile, nil)

	// Act
	hasConsent, err := service.HasValidConsent(context.Background(), userID, "marketing_emails")

	// Assert
	assert.NoError(t, err)
	assert.False(t, hasConsent) // No consent record means no consent

	mockUserRepo.AssertExpectations(t)
}
