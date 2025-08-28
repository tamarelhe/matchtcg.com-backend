package domain

import (
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
