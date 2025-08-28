package domain

import (
	"errors"

	"github.com/google/uuid"
)

var (
	ErrAccessDenied            = errors.New("access denied")
	ErrUserNotGroupMember      = errors.New("user is not a member of the group")
	ErrInsufficientPermissions = errors.New("insufficient permissions")
	ErrEventNotFound           = errors.New("event not found")
	ErrGroupNotFound           = errors.New("group not found")
)

// Permission represents different types of permissions
type Permission string

const (
	PermissionViewEvent        Permission = "view_event"
	PermissionEditEvent        Permission = "edit_event"
	PermissionDeleteEvent      Permission = "delete_event"
	PermissionRSVPToEvent      Permission = "rsvp_to_event"
	PermissionViewAttendees    Permission = "view_attendees"
	PermissionManageGroup      Permission = "manage_group"
	PermissionInviteMembers    Permission = "invite_members"
	PermissionRemoveMembers    Permission = "remove_members"
	PermissionEditGroupDetails Permission = "edit_group_details"
	PermissionDeleteGroup      Permission = "delete_group"
)

// PermissionService handles access control for events and groups
type PermissionService struct{}

// NewPermissionService creates a new PermissionService
func NewPermissionService() *PermissionService {
	return &PermissionService{}
}

// CanViewEvent checks if a user can view an event
func (s *PermissionService) CanViewEvent(event *Event, userID uuid.UUID, userGroups []GroupMember) bool {
	// Event host can always view their own events
	if event.HostUserID == userID {
		return true
	}

	switch event.Visibility {
	case EventVisibilityPublic:
		return true

	case EventVisibilityPrivate:
		// Only the host can view private events
		return event.HostUserID == userID

	case EventVisibilityGroupOnly:
		// Must be a member of the event's group
		if event.GroupID == nil {
			return false
		}
		return s.isUserGroupMember(*event.GroupID, userID, userGroups)

	default:
		return false
	}
}

// CanEditEvent checks if a user can edit an event
func (s *PermissionService) CanEditEvent(event *Event, userID uuid.UUID, userGroups []GroupMember) bool {
	// Event host can always edit their own events
	if event.HostUserID == userID {
		return true
	}

	// Group admins can edit group events
	if event.GroupID != nil {
		return s.hasGroupPermission(*event.GroupID, userID, userGroups, GroupRoleAdmin)
	}

	return false
}

// CanDeleteEvent checks if a user can delete an event
func (s *PermissionService) CanDeleteEvent(event *Event, userID uuid.UUID, userGroups []GroupMember) bool {
	// Event host can always delete their own events
	if event.HostUserID == userID {
		return true
	}

	// Group owners can delete group events
	if event.GroupID != nil {
		return s.hasGroupPermission(*event.GroupID, userID, userGroups, GroupRoleOwner)
	}

	return false
}

// CanRSVPToEvent checks if a user can RSVP to an event
func (s *PermissionService) CanRSVPToEvent(event *Event, userID uuid.UUID, userGroups []GroupMember) bool {
	// Cannot RSVP to your own event
	if event.HostUserID == userID {
		return false
	}

	// Must be able to view the event first
	return s.CanViewEvent(event, userID, userGroups)
}

// CanViewAttendees checks if a user can view event attendees
func (s *PermissionService) CanViewAttendees(event *Event, userID uuid.UUID, userGroups []GroupMember) bool {
	// Event host can always view attendees
	if event.HostUserID == userID {
		return true
	}

	// For public events, anyone can view attendees
	if event.Visibility == EventVisibilityPublic {
		return true
	}

	// For group events, group members can view attendees
	if event.Visibility == EventVisibilityGroupOnly && event.GroupID != nil {
		return s.isUserGroupMember(*event.GroupID, userID, userGroups)
	}

	// For private events, only host can view attendees
	return false
}

// CanManageGroup checks if a user can manage a group (admin-level permissions)
func (s *PermissionService) CanManageGroup(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember) bool {
	return s.hasGroupPermission(groupID, userID, userGroups, GroupRoleAdmin)
}

// CanInviteMembers checks if a user can invite members to a group
func (s *PermissionService) CanInviteMembers(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember) bool {
	return s.hasGroupPermission(groupID, userID, userGroups, GroupRoleAdmin)
}

// CanRemoveMembers checks if a user can remove members from a group
func (s *PermissionService) CanRemoveMembers(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember, targetUserID uuid.UUID) bool {
	// Cannot remove yourself (use leave group instead)
	if userID == targetUserID {
		return false
	}

	userMember := s.getUserGroupMember(groupID, userID, userGroups)
	targetMember := s.getUserGroupMember(groupID, targetUserID, userGroups)

	if userMember == nil || targetMember == nil {
		return false
	}

	// Owners can remove anyone except other owners
	if userMember.Role == GroupRoleOwner {
		return targetMember.Role != GroupRoleOwner
	}

	// Admins can remove regular members only
	if userMember.Role == GroupRoleAdmin {
		return targetMember.Role == GroupRoleMember
	}

	return false
}

// CanEditGroupDetails checks if a user can edit group details
func (s *PermissionService) CanEditGroupDetails(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember) bool {
	return s.hasGroupPermission(groupID, userID, userGroups, GroupRoleAdmin)
}

// CanDeleteGroup checks if a user can delete a group
func (s *PermissionService) CanDeleteGroup(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember) bool {
	return s.hasGroupPermission(groupID, userID, userGroups, GroupRoleOwner)
}

// CanUpdateMemberRole checks if a user can update another member's role
func (s *PermissionService) CanUpdateMemberRole(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember, targetUserID uuid.UUID, newRole GroupRole) bool {
	// Cannot update your own role
	if userID == targetUserID {
		return false
	}

	userMember := s.getUserGroupMember(groupID, userID, userGroups)
	targetMember := s.getUserGroupMember(groupID, targetUserID, userGroups)

	if userMember == nil || targetMember == nil {
		return false
	}

	// Only owners can update roles
	if userMember.Role != GroupRoleOwner {
		return false
	}

	// Cannot change another owner's role
	if targetMember.Role == GroupRoleOwner {
		return false
	}

	// Cannot promote someone to owner (use transfer ownership instead)
	if newRole == GroupRoleOwner {
		return false
	}

	return true
}

// CanTransferOwnership checks if a user can transfer group ownership
func (s *PermissionService) CanTransferOwnership(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember, targetUserID uuid.UUID) bool {
	// Cannot transfer to yourself
	if userID == targetUserID {
		return false
	}

	userMember := s.getUserGroupMember(groupID, userID, userGroups)
	targetMember := s.getUserGroupMember(groupID, targetUserID, userGroups)

	if userMember == nil || targetMember == nil {
		return false
	}

	// Only current owner can transfer ownership
	if userMember.Role != GroupRoleOwner {
		return false
	}

	// Target must be a group member
	return targetMember.Role == GroupRoleAdmin || targetMember.Role == GroupRoleMember
}

// GetUserEventPermissions returns all permissions a user has for an event
func (s *PermissionService) GetUserEventPermissions(event *Event, userID uuid.UUID, userGroups []GroupMember) []Permission {
	var permissions []Permission

	if s.CanViewEvent(event, userID, userGroups) {
		permissions = append(permissions, PermissionViewEvent)
	}

	if s.CanEditEvent(event, userID, userGroups) {
		permissions = append(permissions, PermissionEditEvent)
	}

	if s.CanDeleteEvent(event, userID, userGroups) {
		permissions = append(permissions, PermissionDeleteEvent)
	}

	if s.CanRSVPToEvent(event, userID, userGroups) {
		permissions = append(permissions, PermissionRSVPToEvent)
	}

	if s.CanViewAttendees(event, userID, userGroups) {
		permissions = append(permissions, PermissionViewAttendees)
	}

	return permissions
}

// GetUserGroupPermissions returns all permissions a user has for a group
func (s *PermissionService) GetUserGroupPermissions(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember) []Permission {
	var permissions []Permission

	if s.CanManageGroup(groupID, userID, userGroups) {
		permissions = append(permissions, PermissionManageGroup)
	}

	if s.CanInviteMembers(groupID, userID, userGroups) {
		permissions = append(permissions, PermissionInviteMembers)
	}

	if s.CanEditGroupDetails(groupID, userID, userGroups) {
		permissions = append(permissions, PermissionEditGroupDetails)
	}

	if s.CanDeleteGroup(groupID, userID, userGroups) {
		permissions = append(permissions, PermissionDeleteGroup)
	}

	return permissions
}

// Helper methods

// isUserGroupMember checks if a user is a member of a specific group
func (s *PermissionService) isUserGroupMember(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember) bool {
	for _, member := range userGroups {
		if member.GroupID == groupID && member.UserID == userID {
			return true
		}
	}
	return false
}

// hasGroupPermission checks if a user has at least the specified role in a group
func (s *PermissionService) hasGroupPermission(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember, requiredRole GroupRole) bool {
	member := s.getUserGroupMember(groupID, userID, userGroups)
	if member == nil {
		return false
	}

	return member.HasPermission(requiredRole)
}

// getUserGroupMember returns the GroupMember record for a user in a specific group
func (s *PermissionService) getUserGroupMember(groupID uuid.UUID, userID uuid.UUID, userGroups []GroupMember) *GroupMember {
	for _, member := range userGroups {
		if member.GroupID == groupID && member.UserID == userID {
			return &member
		}
	}
	return nil
}

// FilterVisibleEvents filters a list of events to only include those visible to the user
func (s *PermissionService) FilterVisibleEvents(events []Event, userID uuid.UUID, userGroups []GroupMember) []Event {
	var visibleEvents []Event

	for _, event := range events {
		if s.CanViewEvent(&event, userID, userGroups) {
			visibleEvents = append(visibleEvents, event)
		}
	}

	return visibleEvents
}

// PermissionContext contains context information for permission checks
type PermissionContext struct {
	UserID     uuid.UUID     `json:"user_id"`
	UserGroups []GroupMember `json:"user_groups"`
}

// NewPermissionContext creates a new permission context
func NewPermissionContext(userID uuid.UUID, userGroups []GroupMember) *PermissionContext {
	return &PermissionContext{
		UserID:     userID,
		UserGroups: userGroups,
	}
}
