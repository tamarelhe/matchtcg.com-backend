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
	ErrGroupNotFound           = errors.New("group not found")
	ErrUnauthorizedGroupAccess = errors.New("unauthorized group access")
	ErrCannotDeleteGroup       = errors.New("cannot delete group")
	ErrInvalidGroupOwner       = errors.New("invalid group owner")
	ErrUserNotGroupMember      = errors.New("user is not a group member")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrCannotRemoveOwner       = errors.New("cannot remove group owner")
	ErrInvalidRoleTransition   = errors.New("invalid role transition")
)

// CreateGroupRequest represents the request to create a new group
type CreateGroupRequest struct {
	Name        string    `json:"name" validate:"required,max=100"`
	Description *string   `json:"description,omitempty" validate:"max=1000"`
	OwnerUserID uuid.UUID `json:"owner_user_id" validate:"required"`
}

// CreateGroupResponse represents the response after successful group creation
type CreateGroupResponse struct {
	Group *domain.Group `json:"group"`
}

// UpdateGroupRequest represents the request to update a group
type UpdateGroupRequest struct {
	GroupID     uuid.UUID `json:"group_id" validate:"required"`
	Name        *string   `json:"name,omitempty" validate:"max=100"`
	Description *string   `json:"description,omitempty" validate:"max=1000"`
	UserID      uuid.UUID `json:"user_id" validate:"required"` // User making the request
}

// UpdateGroupResponse represents the response after successful group update
type UpdateGroupResponse struct {
	Group *domain.Group `json:"group"`
}

// DeleteGroupRequest represents the request to delete a group
type DeleteGroupRequest struct {
	GroupID uuid.UUID `json:"group_id" validate:"required"`
	UserID  uuid.UUID `json:"user_id" validate:"required"` // User making the request
}

// CreateGroupUseCase handles group creation
type CreateGroupUseCase struct {
	groupRepo repository.GroupRepository
	userRepo  repository.UserRepository
}

// NewCreateGroupUseCase creates a new CreateGroupUseCase
func NewCreateGroupUseCase(groupRepo repository.GroupRepository, userRepo repository.UserRepository) *CreateGroupUseCase {
	return &CreateGroupUseCase{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// Execute creates a new group with the provided information
func (uc *CreateGroupUseCase) Execute(ctx context.Context, req *CreateGroupRequest) (*CreateGroupResponse, error) {
	// Verify that the owner user exists
	owner, err := uc.userRepo.GetByID(ctx, req.OwnerUserID)
	if err != nil {
		return nil, err
	}
	if owner == nil {
		return nil, ErrInvalidGroupOwner
	}

	// Create group entity
	group := &domain.Group{
		ID:          uuid.New(),
		Name:        req.Name,
		Description: req.Description,
		OwnerUserID: req.OwnerUserID,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
		IsActive:    true,
	}

	// Validate group entity
	if err := group.Validate(); err != nil {
		return nil, err
	}

	// Create group in repository (this also adds the owner as a member)
	if err := uc.groupRepo.Create(ctx, group); err != nil {
		return nil, err
	}

	return &CreateGroupResponse{
		Group: group,
	}, nil
}

// UpdateGroupUseCase handles group updates
type UpdateGroupUseCase struct {
	groupRepo repository.GroupRepository
}

// NewUpdateGroupUseCase creates a new UpdateGroupUseCase
func NewUpdateGroupUseCase(groupRepo repository.GroupRepository) *UpdateGroupUseCase {
	return &UpdateGroupUseCase{
		groupRepo: groupRepo,
	}
}

// Execute updates a group with the provided information
func (uc *UpdateGroupUseCase) Execute(ctx context.Context, req *UpdateGroupRequest) (*UpdateGroupResponse, error) {
	// Get the existing group
	group, err := uc.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	// Check if user has permission to update the group
	canManage, err := uc.groupRepo.CanUserManageGroup(ctx, req.GroupID, req.UserID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, ErrUnauthorizedGroupAccess
	}

	// Update fields if provided
	if req.Name != nil {
		group.Name = *req.Name
	}
	if req.Description != nil {
		group.Description = req.Description
	}

	group.UpdatedAt = time.Now().UTC()

	// Validate updated group
	if err := group.Validate(); err != nil {
		return nil, err
	}

	// Update group in repository
	if err := uc.groupRepo.Update(ctx, group); err != nil {
		return nil, err
	}

	return &UpdateGroupResponse{
		Group: group,
	}, nil
}

// DeleteGroupUseCase handles group deletion
type DeleteGroupUseCase struct {
	groupRepo repository.GroupRepository
}

// NewDeleteGroupUseCase creates a new DeleteGroupUseCase
func NewDeleteGroupUseCase(groupRepo repository.GroupRepository) *DeleteGroupUseCase {
	return &DeleteGroupUseCase{
		groupRepo: groupRepo,
	}
}

// Execute deletes a group
func (uc *DeleteGroupUseCase) Execute(ctx context.Context, req *DeleteGroupRequest) error {
	// Get the existing group
	group, err := uc.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}

	// Check if user is the group owner (only owners can delete groups)
	isOwner, err := uc.groupRepo.IsGroupOwner(ctx, req.GroupID, req.UserID)
	if err != nil {
		return err
	}
	if !isOwner {
		return ErrCannotDeleteGroup
	}

	// Delete group from repository (this will cascade delete members due to foreign key constraints)
	if err := uc.groupRepo.Delete(ctx, req.GroupID); err != nil {
		return err
	}

	return nil
}

// Execute invites a user to join a group with the specified role
func (uc *InviteGroupMemberUseCase) Execute(ctx context.Context, req *InviteGroupMemberRequest) (*InviteGroupMemberResponse, error) {
	// Verify that the group exists
	group, err := uc.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	// Verify that the user to invite exists
	user, err := uc.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user to invite not found")
	}

	// Check if the inviter has permission to manage members
	canManage, err := uc.groupRepo.CanUserManageGroup(ctx, req.GroupID, req.InviterID)
	if err != nil {
		return nil, err
	}
	if !canManage {
		return nil, ErrInsufficientPermissions
	}

	// Check if user is already a member
	isMember, err := uc.groupRepo.IsMember(ctx, req.GroupID, req.UserID)
	if err != nil {
		return nil, err
	}
	if isMember {
		return nil, errors.New("user is already a member of the group")
	}

	// Validate role assignment permissions
	if err := uc.validateRoleAssignment(ctx, req.GroupID, req.InviterID, req.Role); err != nil {
		return nil, err
	}

	// Create group member entity
	member := &domain.GroupMember{
		GroupID:  req.GroupID,
		UserID:   req.UserID,
		Role:     req.Role,
		JoinedAt: time.Now().UTC(),
	}

	// Validate member entity
	if err := member.Validate(); err != nil {
		return nil, err
	}

	// Add member to repository
	if err := uc.groupRepo.AddMember(ctx, member); err != nil {
		return nil, err
	}

	// If assigning owner role, handle ownership transfer
	if req.Role == domain.GroupRoleOwner {
		if err := uc.handleOwnershipTransfer(ctx, req.GroupID, req.InviterID, req.UserID); err != nil {
			return nil, err
		}
	}

	return &InviteGroupMemberResponse{
		Member: member,
	}, nil
}

// validateRoleAssignment checks if the inviter can assign the specified role
func (uc *InviteGroupMemberUseCase) validateRoleAssignment(ctx context.Context, groupID, inviterID uuid.UUID, roleToAssign domain.GroupRole) error {
	inviterRole, err := uc.groupRepo.GetMemberRole(ctx, groupID, inviterID)
	if err != nil {
		return err
	}

	// Only owners can assign owner or admin roles
	if (roleToAssign == domain.GroupRoleOwner || roleToAssign == domain.GroupRoleAdmin) && inviterRole != domain.GroupRoleOwner {
		return ErrInsufficientPermissions
	}

	// Admins can assign member roles
	if roleToAssign == domain.GroupRoleMember && (inviterRole == domain.GroupRoleOwner || inviterRole == domain.GroupRoleAdmin) {
		return nil
	}

	return nil
}

// handleOwnershipTransfer transfers group ownership from current owner to new owner
func (uc *InviteGroupMemberUseCase) handleOwnershipTransfer(ctx context.Context, groupID, currentOwnerID, newOwnerID uuid.UUID) error {
	// Get the group to update ownership
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}

	// Update group owner
	group.OwnerUserID = newOwnerID
	group.UpdatedAt = time.Now().UTC()
	if err := uc.groupRepo.Update(ctx, group); err != nil {
		return err
	}

	// Demote previous owner to admin
	if err := uc.groupRepo.UpdateMemberRole(ctx, groupID, currentOwnerID, domain.GroupRoleAdmin); err != nil {
		return err
	}

	return nil
}

// Execute removes a member from a group
func (uc *RemoveGroupMemberUseCase) Execute(ctx context.Context, req *RemoveGroupMemberRequest) error {
	// Verify that the group exists
	group, err := uc.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return err
	}
	if group == nil {
		return ErrGroupNotFound
	}

	// Check if the user to remove is a member
	member, err := uc.groupRepo.GetMember(ctx, req.GroupID, req.UserID)
	if err != nil {
		return err
	}
	if member == nil {
		return ErrUserNotGroupMember
	}

	// Cannot remove the group owner
	if member.Role == domain.GroupRoleOwner {
		return ErrCannotRemoveOwner
	}

	// Check permissions for removal
	if err := uc.validateRemovalPermissions(ctx, req.GroupID, req.RemoverID, req.UserID); err != nil {
		return err
	}

	// Remove member from repository
	if err := uc.groupRepo.RemoveMember(ctx, req.GroupID, req.UserID); err != nil {
		return err
	}

	return nil
}

// validateRemovalPermissions checks if the remover can remove the specified member
func (uc *RemoveGroupMemberUseCase) validateRemovalPermissions(ctx context.Context, groupID, removerID, userToRemoveID uuid.UUID) error {
	// Users can always remove themselves
	if removerID == userToRemoveID {
		return nil
	}

	// Check if remover has management permissions
	canManage, err := uc.groupRepo.CanUserManageGroup(ctx, groupID, removerID)
	if err != nil {
		return err
	}
	if !canManage {
		return ErrInsufficientPermissions
	}

	// Get roles for additional validation
	removerRole, err := uc.groupRepo.GetMemberRole(ctx, groupID, removerID)
	if err != nil {
		return err
	}

	memberToRemoveRole, err := uc.groupRepo.GetMemberRole(ctx, groupID, userToRemoveID)
	if err != nil {
		return err
	}

	// Admins cannot remove other admins or owners, only owners can
	if removerRole == domain.GroupRoleAdmin && (memberToRemoveRole == domain.GroupRoleAdmin || memberToRemoveRole == domain.GroupRoleOwner) {
		return ErrInsufficientPermissions
	}

	return nil
}

// Execute updates a member's role in a group
func (uc *UpdateMemberRoleUseCase) Execute(ctx context.Context, req *UpdateMemberRoleRequest) (*UpdateMemberRoleResponse, error) {
	// Verify that the group exists
	group, err := uc.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	// Check if the user is a member
	member, err := uc.groupRepo.GetMember(ctx, req.GroupID, req.UserID)
	if err != nil {
		return nil, err
	}
	if member == nil {
		return nil, ErrUserNotGroupMember
	}

	// Validate role update permissions
	if err := uc.validateRoleUpdatePermissions(ctx, req.GroupID, req.UpdaterID, req.UserID, member.Role, req.NewRole); err != nil {
		return nil, err
	}

	// Validate new role
	tempMember := &domain.GroupMember{Role: req.NewRole}
	if err := tempMember.Validate(); err != nil {
		return nil, err
	}

	// Update member role in repository
	if err := uc.groupRepo.UpdateMemberRole(ctx, req.GroupID, req.UserID, req.NewRole); err != nil {
		return nil, err
	}

	// Get updated member
	updatedMember, err := uc.groupRepo.GetMember(ctx, req.GroupID, req.UserID)
	if err != nil {
		return nil, err
	}

	return &UpdateMemberRoleResponse{
		Member: updatedMember,
	}, nil
}

// validateRoleUpdatePermissions checks if the updater can change the member's role
func (uc *UpdateMemberRoleUseCase) validateRoleUpdatePermissions(ctx context.Context, groupID, updaterID, userID uuid.UUID, currentRole, newRole domain.GroupRole) error {
	// Users cannot change their own role
	if updaterID == userID {
		return ErrInsufficientPermissions
	}

	// Get updater's role
	updaterRole, err := uc.groupRepo.GetMemberRole(ctx, groupID, updaterID)
	if err != nil {
		return err
	}

	// Only owners can assign owner or admin roles
	if (newRole == domain.GroupRoleOwner || newRole == domain.GroupRoleAdmin) && updaterRole != domain.GroupRoleOwner {
		return ErrInsufficientPermissions
	}

	// Only owners can change owner roles
	if currentRole == domain.GroupRoleOwner && updaterRole != domain.GroupRoleOwner {
		return ErrInsufficientPermissions
	}

	// Admins cannot change other admin roles
	if updaterRole == domain.GroupRoleAdmin && currentRole == domain.GroupRoleAdmin {
		return ErrInsufficientPermissions
	}

	return nil
}

// handleOwnershipTransfer handles the transfer of group ownership
func (uc *UpdateMemberRoleUseCase) handleOwnershipTransfer(ctx context.Context, groupID, currentOwnerID, newOwnerID uuid.UUID) error {
	// Update the group's owner field
	group, err := uc.groupRepo.GetByID(ctx, groupID)
	if err != nil {
		return err
	}

	group.OwnerUserID = newOwnerID
	group.UpdatedAt = time.Now().UTC()

	if err := uc.groupRepo.Update(ctx, group); err != nil {
		return err
	}

	// Update the new owner's role to owner
	if err := uc.groupRepo.UpdateMemberRole(ctx, groupID, newOwnerID, domain.GroupRoleOwner); err != nil {
		return err
	}

	// Update the previous owner's role to admin
	if err := uc.groupRepo.UpdateMemberRole(ctx, groupID, currentOwnerID, domain.GroupRoleAdmin); err != nil {
		return err
	}

	return nil
}

// Execute retrieves group members with privacy controls
func (uc *GetGroupMembersUseCase) Execute(ctx context.Context, req *GetGroupMembersRequest) (*GetGroupMembersResponse, error) {
	// Verify that the group exists
	group, err := uc.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	// Check if the requester can access the group
	canAccess, err := uc.groupRepo.CanUserAccessGroup(ctx, req.GroupID, req.RequesterID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, ErrUnauthorizedGroupAccess
	}

	// Get group members
	members, err := uc.groupRepo.GetGroupMembers(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	return &GetGroupMembersResponse{
		Members: members,
	}, nil
}

// InviteGroupMemberRequest represents the request to invite a member to a group
type InviteGroupMemberRequest struct {
	GroupID   uuid.UUID        `json:"group_id" validate:"required"`
	UserID    uuid.UUID        `json:"user_id" validate:"required"`    // User to invite
	Role      domain.GroupRole `json:"role" validate:"required"`       // Role to assign
	InviterID uuid.UUID        `json:"inviter_id" validate:"required"` // User making the invitation
}

// InviteGroupMemberResponse represents the response after successful member invitation
type InviteGroupMemberResponse struct {
	Member *domain.GroupMember `json:"member"`
}

// RemoveGroupMemberRequest represents the request to remove a member from a group
type RemoveGroupMemberRequest struct {
	GroupID   uuid.UUID `json:"group_id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`    // User to remove
	RemoverID uuid.UUID `json:"remover_id" validate:"required"` // User making the removal
}

// UpdateMemberRoleRequest represents the request to update a member's role
type UpdateMemberRoleRequest struct {
	GroupID   uuid.UUID        `json:"group_id" validate:"required"`
	UserID    uuid.UUID        `json:"user_id" validate:"required"`    // User whose role to update
	NewRole   domain.GroupRole `json:"new_role" validate:"required"`   // New role to assign
	UpdaterID uuid.UUID        `json:"updater_id" validate:"required"` // User making the update
}

// UpdateMemberRoleResponse represents the response after successful role update
type UpdateMemberRoleResponse struct {
	Member *domain.GroupMember `json:"member"`
}

// GetGroupMembersRequest represents the request to get group members
type GetGroupMembersRequest struct {
	GroupID     uuid.UUID `json:"group_id" validate:"required"`
	RequesterID uuid.UUID `json:"requester_id" validate:"required"` // User making the request
}

// GetGroupMembersResponse represents the response with group members
type GetGroupMembersResponse struct {
	Members []*domain.GroupMember `json:"members"`
}

// InviteGroupMemberUseCase handles group member invitations
type InviteGroupMemberUseCase struct {
	groupRepo repository.GroupRepository
	userRepo  repository.UserRepository
}

// NewInviteGroupMemberUseCase creates a new InviteGroupMemberUseCase
func NewInviteGroupMemberUseCase(groupRepo repository.GroupRepository, userRepo repository.UserRepository) *InviteGroupMemberUseCase {
	return &InviteGroupMemberUseCase{
		groupRepo: groupRepo,
		userRepo:  userRepo,
	}
}

// RemoveGroupMemberUseCase handles group member removal
type RemoveGroupMemberUseCase struct {
	groupRepo repository.GroupRepository
}

// NewRemoveGroupMemberUseCase creates a new RemoveGroupMemberUseCase
func NewRemoveGroupMemberUseCase(groupRepo repository.GroupRepository) *RemoveGroupMemberUseCase {
	return &RemoveGroupMemberUseCase{
		groupRepo: groupRepo,
	}
}

// UpdateMemberRoleUseCase handles member role updates
type UpdateMemberRoleUseCase struct {
	groupRepo repository.GroupRepository
}

// NewUpdateMemberRoleUseCase creates a new UpdateMemberRoleUseCase
func NewUpdateMemberRoleUseCase(groupRepo repository.GroupRepository) *UpdateMemberRoleUseCase {
	return &UpdateMemberRoleUseCase{
		groupRepo: groupRepo,
	}
}

// GetGroupMembersUseCase handles retrieving group members
type GetGroupMembersUseCase struct {
	groupRepo repository.GroupRepository
}

// NewGetGroupMembersUseCase creates a new GetGroupMembersUseCase
func NewGetGroupMembersUseCase(groupRepo repository.GroupRepository) *GetGroupMembersUseCase {
	return &GetGroupMembersUseCase{
		groupRepo: groupRepo,
	}
}

// Execute retrieves group events with privacy controls
func (uc *GetGroupEventsUseCase) Execute(ctx context.Context, req *GetGroupEventsRequest) (*GetGroupEventsResponse, error) {
	// Verify that the group exists
	group, err := uc.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, ErrGroupNotFound
	}

	// Check if requester can access the group
	canAccess, err := uc.groupRepo.CanUserAccessGroup(ctx, req.GroupID, req.RequesterID)
	if err != nil {
		return nil, err
	}
	if !canAccess {
		return nil, ErrUnauthorizedGroupAccess
	}

	// Get group events
	events, err := uc.eventRepo.GetGroupEvents(ctx, req.GroupID, req.Limit, req.Offset)
	if err != nil {
		return nil, err
	}

	return &GetGroupEventsResponse{
		Events: events,
	}, nil
}

// GetGroupEventsRequest represents the request to get events for a group
type GetGroupEventsRequest struct {
	GroupID     uuid.UUID `json:"group_id" validate:"required"`
	RequesterID uuid.UUID `json:"requester_id" validate:"required"` // User making the request
	Limit       int       `json:"limit" validate:"min=1,max=100"`
	Offset      int       `json:"offset" validate:"min=0"`
}

// GetGroupEventsResponse represents the response with group events
type GetGroupEventsResponse struct {
	Events []*domain.Event `json:"events"`
}

// GetGroupEventsUseCase handles retrieving events for a specific group
type GetGroupEventsUseCase struct {
	groupRepo repository.GroupRepository
	eventRepo repository.EventRepository
}

// NewGetGroupEventsUseCase creates a new GetGroupEventsUseCase
func NewGetGroupEventsUseCase(groupRepo repository.GroupRepository, eventRepo repository.EventRepository) *GetGroupEventsUseCase {
	return &GetGroupEventsUseCase{
		groupRepo: groupRepo,
		eventRepo: eventRepo,
	}
}

// GroupManagementUseCase provides a unified interface for all group management operations
type GroupManagementUseCase struct {
	createGroupUseCase       *CreateGroupUseCase
	updateGroupUseCase       *UpdateGroupUseCase
	deleteGroupUseCase       *DeleteGroupUseCase
	inviteGroupMemberUseCase *InviteGroupMemberUseCase
	removeGroupMemberUseCase *RemoveGroupMemberUseCase
	updateMemberRoleUseCase  *UpdateMemberRoleUseCase
	getGroupMembersUseCase   *GetGroupMembersUseCase
	getGroupEventsUseCase    *GetGroupEventsUseCase
}

// NewGroupManagementUseCase creates a new unified group management use case
func NewGroupManagementUseCase(
	groupRepo repository.GroupRepository,
	userRepo repository.UserRepository,
	eventRepo repository.EventRepository,
) *GroupManagementUseCase {
	return &GroupManagementUseCase{
		createGroupUseCase:       NewCreateGroupUseCase(groupRepo, userRepo),
		updateGroupUseCase:       NewUpdateGroupUseCase(groupRepo),
		deleteGroupUseCase:       NewDeleteGroupUseCase(groupRepo),
		inviteGroupMemberUseCase: NewInviteGroupMemberUseCase(groupRepo, userRepo),
		removeGroupMemberUseCase: NewRemoveGroupMemberUseCase(groupRepo),
		updateMemberRoleUseCase:  NewUpdateMemberRoleUseCase(groupRepo),
		getGroupMembersUseCase:   NewGetGroupMembersUseCase(groupRepo),
		getGroupEventsUseCase:    NewGetGroupEventsUseCase(groupRepo, eventRepo),
	}
}

// CreateGroup creates a new group
func (uc *GroupManagementUseCase) CreateGroup(ctx context.Context, req *CreateGroupRequest) (*CreateGroupResponse, error) {
	return uc.createGroupUseCase.Execute(ctx, req)
}

// UpdateGroup updates an existing group
func (uc *GroupManagementUseCase) UpdateGroup(ctx context.Context, req *UpdateGroupRequest) (*UpdateGroupResponse, error) {
	return uc.updateGroupUseCase.Execute(ctx, req)
}

// DeleteGroup deletes a group
func (uc *GroupManagementUseCase) DeleteGroup(ctx context.Context, req *DeleteGroupRequest) error {
	return uc.deleteGroupUseCase.Execute(ctx, req)
}

// InviteGroupMember invites a user to join a group
func (uc *GroupManagementUseCase) InviteGroupMember(ctx context.Context, req *InviteGroupMemberRequest) (*InviteGroupMemberResponse, error) {
	return uc.inviteGroupMemberUseCase.Execute(ctx, req)
}

// RemoveGroupMember removes a member from a group
func (uc *GroupManagementUseCase) RemoveGroupMember(ctx context.Context, req *RemoveGroupMemberRequest) error {
	return uc.removeGroupMemberUseCase.Execute(ctx, req)
}

// UpdateMemberRole updates a member's role in a group
func (uc *GroupManagementUseCase) UpdateMemberRole(ctx context.Context, req *UpdateMemberRoleRequest) (*UpdateMemberRoleResponse, error) {
	return uc.updateMemberRoleUseCase.Execute(ctx, req)
}

// GetGroupMembers retrieves group members
func (uc *GroupManagementUseCase) GetGroupMembers(ctx context.Context, req *GetGroupMembersRequest) (*GetGroupMembersResponse, error) {
	return uc.getGroupMembersUseCase.Execute(ctx, req)
}

// GetGroupEvents retrieves events for a group
func (uc *GroupManagementUseCase) GetGroupEvents(ctx context.Context, req *GetGroupEventsRequest) (*GetGroupEventsResponse, error) {
	return uc.getGroupEventsUseCase.Execute(ctx, req)
}

// GetGroupRequest represents the request to get a group
type GetGroupRequest struct {
	GroupID          uuid.UUID  `json:"group_id" validate:"required"`
	RequestingUserID *uuid.UUID `json:"requesting_user_id,omitempty"`
	IncludeMembers   bool       `json:"include_members,omitempty"`
}

// GetGroup retrieves a group by ID with optional member information
func (uc *GroupManagementUseCase) GetGroup(ctx context.Context, req *GetGroupRequest) (*domain.GroupWithDetails, error) {
	// Get group
	group, err := uc.createGroupUseCase.groupRepo.GetByID(ctx, req.GroupID)
	if err != nil {
		return nil, err
	}

	// Create response with basic group info
	groupWithDetails := &domain.GroupWithDetails{
		Group: *group,
	}

	// Get member count and members
	members, err := uc.createGroupUseCase.groupRepo.GetGroupMembers(ctx, req.GroupID)
	if err == nil {
		groupWithDetails.MemberCount = len(members)

		// Include members if requested
		if req.IncludeMembers {
			groupWithDetails.Members = members
		}

		// Get user's role if authenticated
		if req.RequestingUserID != nil {
			for _, member := range members {
				if member.UserID == *req.RequestingUserID {
					userRole := string(member.Role)
					groupWithDetails.UserRole = &userRole
					break
				}
			}
		}
	}

	return groupWithDetails, nil
}

// GetUserGroupsRequest represents the request to get user's groups
type GetUserGroupsRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
}

// GetUserGroupsResponse represents the response with user's groups
type GetUserGroupsResponse struct {
	Groups []*domain.GroupWithDetails `json:"groups"`
}

// GetUserGroups retrieves all groups for a user
func (uc *GroupManagementUseCase) GetUserGroups(ctx context.Context, req *GetUserGroupsRequest) (*GetUserGroupsResponse, error) {
	// Get user's groups
	userGroups, err := uc.createGroupUseCase.groupRepo.GetUserGroups(ctx, req.UserID)
	if err != nil {
		return nil, err
	}

	var groups []*domain.GroupWithDetails
	for _, group := range userGroups {
		// Get member count
		memberCount, err := uc.createGroupUseCase.groupRepo.GetMemberCount(ctx, group.ID)
		if err != nil {
			memberCount = 0 // Default to 0 if we can't get the count
		}

		// Get user's role in this group
		userRole, err := uc.createGroupUseCase.groupRepo.GetMemberRole(ctx, group.ID, req.UserID)
		var userRoleStr *string
		if err == nil {
			roleStr := string(userRole)
			userRoleStr = &roleStr
		}

		// Create group with details
		groupWithDetails := &domain.GroupWithDetails{
			Group:       *group,
			MemberCount: memberCount,
			UserRole:    userRoleStr,
		}

		groups = append(groups, groupWithDetails)
	}

	return &GetUserGroupsResponse{
		Groups: groups,
	}, nil
}
