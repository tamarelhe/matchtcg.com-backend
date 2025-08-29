package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
)

// ExportUserDataRequest represents the request to export user data
type ExportUserDataRequest struct {
	UserID uuid.UUID `json:"-"` // Set from authentication context
}

// ExportUserDataResponse represents the complete user data export
type ExportUserDataResponse struct {
	ExportedAt time.Time              `json:"exported_at"`
	UserID     uuid.UUID              `json:"user_id"`
	Data       map[string]interface{} `json:"data"`
}

// DeleteUserAccountRequest represents the request to delete user account
type DeleteUserAccountRequest struct {
	UserID uuid.UUID `json:"-"` // Set from authentication context
}

// DeleteUserAccountResponse represents the response after account deletion
type DeleteUserAccountResponse struct {
	DeletedAt time.Time `json:"deleted_at"`
	UserID    uuid.UUID `json:"user_id"`
	Message   string    `json:"message"`
}

// ConsentUpdateRequest represents a request to update user consent
type ConsentUpdateRequest struct {
	UserID      uuid.UUID              `json:"-"` // Set from authentication context
	ConsentType string                 `json:"consent_type" validate:"required"`
	Granted     bool                   `json:"granted"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// ConsentRecord represents a consent record
type ConsentRecord struct {
	ID          uuid.UUID              `json:"id"`
	UserID      uuid.UUID              `json:"user_id"`
	ConsentType string                 `json:"consent_type"`
	Granted     bool                   `json:"granted"`
	GrantedAt   time.Time              `json:"granted_at"`
	RevokedAt   *time.Time             `json:"revoked_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// ExportUserDataUseCase handles complete user data export for GDPR compliance
type ExportUserDataUseCase struct {
	userRepo         repository.UserRepository
	eventRepo        repository.EventRepository
	groupRepo        repository.GroupRepository
	notificationRepo repository.NotificationRepository
}

// NewExportUserDataUseCase creates a new ExportUserDataUseCase
func NewExportUserDataUseCase(
	userRepo repository.UserRepository,
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	notificationRepo repository.NotificationRepository,
) *ExportUserDataUseCase {
	return &ExportUserDataUseCase{
		userRepo:         userRepo,
		eventRepo:        eventRepo,
		groupRepo:        groupRepo,
		notificationRepo: notificationRepo,
	}
}

// Execute exports all user data in a machine-readable format
func (uc *ExportUserDataUseCase) Execute(ctx context.Context, req *ExportUserDataRequest) (*ExportUserDataResponse, error) {
	// Get user with profile
	userWithProfile, err := uc.userRepo.GetUserWithProfile(ctx, req.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Initialize export data structure
	exportData := make(map[string]interface{})

	// Export user account data
	exportData["account"] = map[string]interface{}{
		"id":         userWithProfile.User.ID,
		"email":      userWithProfile.User.Email,
		"created_at": userWithProfile.User.CreatedAt,
		"updated_at": userWithProfile.User.UpdatedAt,
		"is_active":  userWithProfile.User.IsActive,
		"last_login": userWithProfile.User.LastLogin,
	}

	// Export profile data
	if userWithProfile.Profile != nil {
		exportData["profile"] = userWithProfile.Profile
	}

	// Export user's events (hosted events)
	userEvents, err := uc.eventRepo.GetUserEvents(ctx, req.UserID, 1000, 0) // Large limit for export
	if err == nil {
		exportData["hosted_events"] = userEvents
	}

	// Export user's RSVPs
	userRSVPs, err := uc.eventRepo.GetUserRSVPs(ctx, req.UserID)
	if err == nil {
		exportData["event_rsvps"] = userRSVPs
	}

	// Export user's groups
	userGroups, err := uc.groupRepo.GetUserGroups(ctx, req.UserID)
	if err == nil {
		exportData["groups"] = userGroups
	}

	// Export user's group memberships
	userGroupMemberships := make([]map[string]interface{}, 0)
	for _, group := range userGroups {
		member, err := uc.groupRepo.GetMember(ctx, group.ID, req.UserID)
		if err == nil {
			userGroupMemberships = append(userGroupMemberships, map[string]interface{}{
				"group_id":  group.ID,
				"role":      member.Role,
				"joined_at": member.JoinedAt,
			})
		}
	}
	exportData["group_memberships"] = userGroupMemberships

	// Export user's notifications
	userNotifications, err := uc.notificationRepo.GetUserNotifications(ctx, req.UserID, 1000, 0)
	if err == nil {
		exportData["notifications"] = userNotifications
	}

	// Export additional data from repository
	additionalData, err := uc.userRepo.ExportUserData(ctx, req.UserID)
	if err == nil {
		for key, value := range additionalData {
			exportData[key] = value
		}
	}

	return &ExportUserDataResponse{
		ExportedAt: time.Now().UTC(),
		UserID:     req.UserID,
		Data:       exportData,
	}, nil
}

// DeleteUserAccountUseCase handles complete user account deletion with cascading cleanup
type DeleteUserAccountUseCase struct {
	userRepo         repository.UserRepository
	eventRepo        repository.EventRepository
	groupRepo        repository.GroupRepository
	notificationRepo repository.NotificationRepository
}

// NewDeleteUserAccountUseCase creates a new DeleteUserAccountUseCase
func NewDeleteUserAccountUseCase(
	userRepo repository.UserRepository,
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	notificationRepo repository.NotificationRepository,
) *DeleteUserAccountUseCase {
	return &DeleteUserAccountUseCase{
		userRepo:         userRepo,
		eventRepo:        eventRepo,
		groupRepo:        groupRepo,
		notificationRepo: notificationRepo,
	}
}

// Execute permanently deletes user account and all associated data
func (uc *DeleteUserAccountUseCase) Execute(ctx context.Context, req *DeleteUserAccountRequest) (*DeleteUserAccountResponse, error) {
	// Verify user exists
	_, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	// Get user's groups where they are the owner
	ownedGroups, err := uc.groupRepo.GetGroupsByOwner(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	// Handle group ownership transfer or deletion
	for _, group := range ownedGroups {
		// Get group members to potentially transfer ownership
		members, err := uc.groupRepo.GetGroupMembers(ctx, group.ID)
		if err != nil {
			continue
		}

		// Find an admin to transfer ownership to
		var newOwner *domain.GroupMember
		for _, member := range members {
			if member.UserID != req.UserID && member.Role == domain.GroupRoleAdmin {
				newOwner = member
				break
			}
		}

		if newOwner != nil {
			// Transfer ownership to an admin
			group.OwnerUserID = newOwner.UserID
			group.UpdatedAt = time.Now().UTC()
			uc.groupRepo.Update(ctx, group)
			uc.groupRepo.UpdateMemberRole(ctx, group.ID, newOwner.UserID, domain.GroupRoleOwner)
		} else {
			// No suitable admin found, delete the group
			uc.groupRepo.Delete(ctx, group.ID)
		}
	}

	// Get user's hosted events
	userEvents, err := uc.eventRepo.GetUserEvents(ctx, req.UserID, 1000, 0)
	if err == nil {
		// Delete or transfer hosted events
		for _, event := range userEvents {
			// For MVP, we'll delete the events
			// In production, you might want to transfer to group admins or mark as cancelled
			uc.eventRepo.Delete(ctx, event.ID)
		}
	}

	// Remove user from all groups
	userGroups, err := uc.groupRepo.GetUserGroups(ctx, req.UserID)
	if err == nil {
		for _, group := range userGroups {
			uc.groupRepo.RemoveMember(ctx, group.ID, req.UserID)
		}
	}

	// Delete user's RSVPs
	userRSVPs, err := uc.eventRepo.GetUserRSVPs(ctx, req.UserID)
	if err == nil {
		for _, rsvp := range userRSVPs {
			uc.eventRepo.DeleteRSVP(ctx, rsvp.EventID, req.UserID)
		}
	}

	// Delete user's notifications
	// This would typically be handled by the repository's DeleteUserData method

	// Finally, delete all user data (this should cascade to profile and other related data)
	if err := uc.userRepo.DeleteUserData(ctx, req.UserID); err != nil {
		return nil, err
	}

	return &DeleteUserAccountResponse{
		DeletedAt: time.Now().UTC(),
		UserID:    req.UserID,
		Message:   "User account and all associated data have been permanently deleted",
	}, nil
}

// ConsentManagementService handles user consent tracking and management
type ConsentManagementService struct {
	userRepo repository.UserRepository
	// In a real implementation, you might have a dedicated ConsentRepository
	// For now, we'll store consent data in the user's profile or a separate table
}

// NewConsentManagementService creates a new ConsentManagementService
func NewConsentManagementService(userRepo repository.UserRepository) *ConsentManagementService {
	return &ConsentManagementService{
		userRepo: userRepo,
	}
}

// UpdateConsent updates or creates a consent record for a user
func (s *ConsentManagementService) UpdateConsent(ctx context.Context, req *ConsentUpdateRequest) (*ConsentRecord, error) {
	// Get user profile to store consent information
	profile, err := s.userRepo.GetProfile(ctx, req.UserID)
	if err != nil {
		return nil, ErrProfileNotFound
	}

	// Initialize consent data if not exists
	if profile.CommunicationPreferences == nil {
		profile.CommunicationPreferences = make(map[string]interface{})
	}

	// Create consent record
	consentRecord := &ConsentRecord{
		ID:          uuid.New(),
		UserID:      req.UserID,
		ConsentType: req.ConsentType,
		Granted:     req.Granted,
		GrantedAt:   time.Now().UTC(),
		Metadata:    req.Metadata,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if !req.Granted {
		now := time.Now().UTC()
		consentRecord.RevokedAt = &now
	}

	// Store consent in profile (in a real implementation, this would be in a dedicated consent table)
	consentKey := "consent_" + req.ConsentType
	consentData, _ := json.Marshal(consentRecord)
	var consentMap map[string]interface{}
	json.Unmarshal(consentData, &consentMap)

	profile.CommunicationPreferences[consentKey] = consentMap
	profile.UpdatedAt = time.Now().UTC()

	// Update profile
	if err := s.userRepo.UpdateProfile(ctx, profile); err != nil {
		return nil, err
	}

	return consentRecord, nil
}

// GetUserConsents retrieves all consent records for a user
func (s *ConsentManagementService) GetUserConsents(ctx context.Context, userID uuid.UUID) ([]*ConsentRecord, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, ErrProfileNotFound
	}

	var consents []*ConsentRecord

	if profile.CommunicationPreferences != nil {
		for key, value := range profile.CommunicationPreferences {
			if len(key) > 8 && key[:8] == "consent_" {
				consentData, _ := json.Marshal(value)
				var consent ConsentRecord
				if json.Unmarshal(consentData, &consent) == nil {
					consents = append(consents, &consent)
				}
			}
		}
	}

	return consents, nil
}

// GetConsent retrieves a specific consent record for a user
func (s *ConsentManagementService) GetConsent(ctx context.Context, userID uuid.UUID, consentType string) (*ConsentRecord, error) {
	profile, err := s.userRepo.GetProfile(ctx, userID)
	if err != nil {
		return nil, ErrProfileNotFound
	}

	if profile.CommunicationPreferences == nil {
		return nil, errors.New("consent not found")
	}

	consentKey := "consent_" + consentType
	consentValue, exists := profile.CommunicationPreferences[consentKey]
	if !exists {
		return nil, errors.New("consent not found")
	}

	consentData, _ := json.Marshal(consentValue)
	var consent ConsentRecord
	if err := json.Unmarshal(consentData, &consent); err != nil {
		return nil, err
	}

	return &consent, nil
}

// HasValidConsent checks if user has granted consent for a specific type
func (s *ConsentManagementService) HasValidConsent(ctx context.Context, userID uuid.UUID, consentType string) (bool, error) {
	consent, err := s.GetConsent(ctx, userID, consentType)
	if err != nil {
		return false, nil // No consent record means no consent
	}

	return consent.Granted && consent.RevokedAt == nil, nil
}

// GDPRComplianceUseCase provides a unified interface for GDPR compliance operations
type GDPRComplianceUseCase struct {
	exportUseCase  *ExportUserDataUseCase
	deleteUseCase  *DeleteUserAccountUseCase
	consentService *ConsentManagementService
}

// NewGDPRComplianceUseCase creates a new unified GDPR compliance use case
func NewGDPRComplianceUseCase(
	userRepo repository.UserRepository,
	eventRepo repository.EventRepository,
	groupRepo repository.GroupRepository,
	notificationRepo repository.NotificationRepository,
) *GDPRComplianceUseCase {
	return &GDPRComplianceUseCase{
		exportUseCase:  NewExportUserDataUseCase(userRepo, eventRepo, groupRepo, notificationRepo),
		deleteUseCase:  NewDeleteUserAccountUseCase(userRepo, eventRepo, groupRepo, notificationRepo),
		consentService: NewConsentManagementService(userRepo),
	}
}

// ExportUserData exports all user data for GDPR compliance
func (uc *GDPRComplianceUseCase) ExportUserData(ctx context.Context, userID uuid.UUID) (map[string]interface{}, error) {
	req := &ExportUserDataRequest{
		UserID: userID,
	}

	result, err := uc.exportUseCase.Execute(ctx, req)
	if err != nil {
		return nil, err
	}

	return result.Data, nil
}

// DeleteUserAccount permanently deletes user account and all associated data
func (uc *GDPRComplianceUseCase) DeleteUserAccount(ctx context.Context, userID uuid.UUID) error {
	req := &DeleteUserAccountRequest{
		UserID: userID,
	}

	_, err := uc.deleteUseCase.Execute(ctx, req)
	return err
}

// UpdateConsent updates user consent for a specific type
func (uc *GDPRComplianceUseCase) UpdateConsent(ctx context.Context, userID uuid.UUID, consentType string, granted bool, metadata map[string]interface{}) (*ConsentRecord, error) {
	req := &ConsentUpdateRequest{
		UserID:      userID,
		ConsentType: consentType,
		Granted:     granted,
		Metadata:    metadata,
	}

	return uc.consentService.UpdateConsent(ctx, req)
}

// GetUserConsents retrieves all consent records for a user
func (uc *GDPRComplianceUseCase) GetUserConsents(ctx context.Context, userID uuid.UUID) ([]*ConsentRecord, error) {
	return uc.consentService.GetUserConsents(ctx, userID)
}

// HasValidConsent checks if user has granted consent for a specific type
func (uc *GDPRComplianceUseCase) HasValidConsent(ctx context.Context, userID uuid.UUID, consentType string) (bool, error) {
	return uc.consentService.HasValidConsent(ctx, userID, consentType)
}
