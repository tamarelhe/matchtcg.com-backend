package domain

import (
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
)

// GroupRole represents the role of a user in a group
type GroupRole string

const (
	GroupRoleOwner  GroupRole = "owner"
	GroupRoleAdmin  GroupRole = "admin"
	GroupRoleMember GroupRole = "member"
)

// Group represents a group in the system
type Group struct {
	ID          uuid.UUID `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	OwnerUserID uuid.UUID `json:"owner_user_id" db:"owner_user_id"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
	IsActive    bool      `json:"is_active" db:"is_active"`
}

// GroupMember represents a member of a group
type GroupMember struct {
	GroupID  uuid.UUID `json:"group_id" db:"group_id"`
	UserID   uuid.UUID `json:"user_id" db:"user_id"`
	Role     GroupRole `json:"role" db:"role"`
	JoinedAt time.Time `json:"joined_at" db:"joined_at"`
}

// GroupWithMembers represents a group with its members
type GroupWithMembers struct {
	Group
	Members []GroupMember `json:"members,omitempty"`
}

var (
	ErrEmptyGroupName          = errors.New("group name cannot be empty")
	ErrGroupNameTooLong        = errors.New("group name cannot exceed 100 characters")
	ErrGroupDescriptionTooLong = errors.New("group description cannot exceed 1000 characters")
	ErrInvalidGroupRole        = errors.New("invalid group role")
)

// Validate validates the Group entity
func (g *Group) Validate() error {
	if strings.TrimSpace(g.Name) == "" {
		return ErrEmptyGroupName
	}

	if len(g.Name) > 100 {
		return ErrGroupNameTooLong
	}

	if g.Description != nil && len(*g.Description) > 1000 {
		return ErrGroupDescriptionTooLong
	}

	return nil
}

// Validate validates the GroupMember entity
func (gm *GroupMember) Validate() error {
	if !gm.IsValidRole() {
		return ErrInvalidGroupRole
	}
	return nil
}

// IsValidRole checks if the group role is valid
func (gm *GroupMember) IsValidRole() bool {
	switch gm.Role {
	case GroupRoleOwner, GroupRoleAdmin, GroupRoleMember:
		return true
	default:
		return false
	}
}

// IsOwner checks if the member is the group owner
func (gm *GroupMember) IsOwner() bool {
	return gm.Role == GroupRoleOwner
}

// IsAdmin checks if the member is an admin (owner or admin role)
func (gm *GroupMember) IsAdmin() bool {
	return gm.Role == GroupRoleOwner || gm.Role == GroupRoleAdmin
}

// CanManageMembers checks if the member can manage other members
func (gm *GroupMember) CanManageMembers() bool {
	return gm.IsAdmin()
}

// CanEditGroup checks if the member can edit group details
func (gm *GroupMember) CanEditGroup() bool {
	return gm.IsAdmin()
}

// CanDeleteGroup checks if the member can delete the group
func (gm *GroupMember) CanDeleteGroup() bool {
	return gm.IsOwner()
}

// HasPermission checks if the member has a specific permission level
func (gm *GroupMember) HasPermission(requiredRole GroupRole) bool {
	switch requiredRole {
	case GroupRoleMember:
		return true // All members have member permissions
	case GroupRoleAdmin:
		return gm.Role == GroupRoleAdmin || gm.Role == GroupRoleOwner
	case GroupRoleOwner:
		return gm.Role == GroupRoleOwner
	default:
		return false
	}
}

// GroupWithDetails represents a group with additional details
type GroupWithDetails struct {
	Group
	Owner       *UserWithProfile `json:"owner,omitempty"`
	Members     []*GroupMember   `json:"members,omitempty"`
	MemberCount int              `json:"member_count"`
	UserRole    *string          `json:"user_role,omitempty"`
}
