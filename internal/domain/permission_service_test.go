package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestPermissionService_CanViewEvent(t *testing.T) {
	service := NewPermissionService()

	hostID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	userGroups := []GroupMember{
		{GroupID: groupID, UserID: userID, Role: GroupRoleMember, JoinedAt: time.Now()},
	}

	tests := []struct {
		name       string
		event      Event
		userID     uuid.UUID
		userGroups []GroupMember
		want       bool
	}{
		{
			name: "host can view own event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPrivate,
			},
			userID:     hostID,
			userGroups: []GroupMember{},
			want:       true,
		},
		{
			name: "anyone can view public event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPublic,
			},
			userID:     userID,
			userGroups: []GroupMember{},
			want:       true,
		},
		{
			name: "non-host cannot view private event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPrivate,
			},
			userID:     userID,
			userGroups: []GroupMember{},
			want:       false,
		},
		{
			name: "group member can view group-only event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				GroupID:    &groupID,
				Visibility: EventVisibilityGroupOnly,
			},
			userID:     userID,
			userGroups: userGroups,
			want:       true,
		},
		{
			name: "non-group member cannot view group-only event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				GroupID:    &groupID,
				Visibility: EventVisibilityGroupOnly,
			},
			userID:     userID,
			userGroups: []GroupMember{}, // Not a member
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanViewEvent(&tt.event, tt.userID, tt.userGroups)
			if got != tt.want {
				t.Errorf("CanViewEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionService_CanEditEvent(t *testing.T) {
	service := NewPermissionService()

	hostID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	userGroups := []GroupMember{
		{GroupID: groupID, UserID: userID, Role: GroupRoleAdmin, JoinedAt: time.Now()},
	}

	tests := []struct {
		name       string
		event      Event
		userID     uuid.UUID
		userGroups []GroupMember
		want       bool
	}{
		{
			name: "host can edit own event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPublic,
			},
			userID:     hostID,
			userGroups: []GroupMember{},
			want:       true,
		},
		{
			name: "group admin can edit group event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				GroupID:    &groupID,
				Visibility: EventVisibilityGroupOnly,
			},
			userID:     userID,
			userGroups: userGroups,
			want:       true,
		},
		{
			name: "regular user cannot edit other's event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPublic,
			},
			userID:     userID,
			userGroups: []GroupMember{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanEditEvent(&tt.event, tt.userID, tt.userGroups)
			if got != tt.want {
				t.Errorf("CanEditEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionService_CanDeleteEvent(t *testing.T) {
	service := NewPermissionService()

	hostID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	ownerGroups := []GroupMember{
		{GroupID: groupID, UserID: userID, Role: GroupRoleOwner, JoinedAt: time.Now()},
	}

	adminGroups := []GroupMember{
		{GroupID: groupID, UserID: userID, Role: GroupRoleAdmin, JoinedAt: time.Now()},
	}

	tests := []struct {
		name       string
		event      Event
		userID     uuid.UUID
		userGroups []GroupMember
		want       bool
	}{
		{
			name: "host can delete own event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPublic,
			},
			userID:     hostID,
			userGroups: []GroupMember{},
			want:       true,
		},
		{
			name: "group owner can delete group event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				GroupID:    &groupID,
				Visibility: EventVisibilityGroupOnly,
			},
			userID:     userID,
			userGroups: ownerGroups,
			want:       true,
		},
		{
			name: "group admin cannot delete group event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				GroupID:    &groupID,
				Visibility: EventVisibilityGroupOnly,
			},
			userID:     userID,
			userGroups: adminGroups,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanDeleteEvent(&tt.event, tt.userID, tt.userGroups)
			if got != tt.want {
				t.Errorf("CanDeleteEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionService_CanRSVPToEvent(t *testing.T) {
	service := NewPermissionService()

	hostID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		event      Event
		userID     uuid.UUID
		userGroups []GroupMember
		want       bool
	}{
		{
			name: "user can RSVP to public event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPublic,
			},
			userID:     userID,
			userGroups: []GroupMember{},
			want:       true,
		},
		{
			name: "host cannot RSVP to own event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPublic,
			},
			userID:     hostID,
			userGroups: []GroupMember{},
			want:       false,
		},
		{
			name: "user cannot RSVP to private event they cannot view",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPrivate,
			},
			userID:     userID,
			userGroups: []GroupMember{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanRSVPToEvent(&tt.event, tt.userID, tt.userGroups)
			if got != tt.want {
				t.Errorf("CanRSVPToEvent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionService_CanViewAttendees(t *testing.T) {
	service := NewPermissionService()

	hostID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	userGroups := []GroupMember{
		{GroupID: groupID, UserID: userID, Role: GroupRoleMember, JoinedAt: time.Now()},
	}

	tests := []struct {
		name       string
		event      Event
		userID     uuid.UUID
		userGroups []GroupMember
		want       bool
	}{
		{
			name: "host can view attendees",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPrivate,
			},
			userID:     hostID,
			userGroups: []GroupMember{},
			want:       true,
		},
		{
			name: "anyone can view attendees of public event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPublic,
			},
			userID:     userID,
			userGroups: []GroupMember{},
			want:       true,
		},
		{
			name: "group member can view attendees of group event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				GroupID:    &groupID,
				Visibility: EventVisibilityGroupOnly,
			},
			userID:     userID,
			userGroups: userGroups,
			want:       true,
		},
		{
			name: "non-host cannot view attendees of private event",
			event: Event{
				ID:         uuid.New(),
				HostUserID: hostID,
				Visibility: EventVisibilityPrivate,
			},
			userID:     userID,
			userGroups: []GroupMember{},
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanViewAttendees(&tt.event, tt.userID, tt.userGroups)
			if got != tt.want {
				t.Errorf("CanViewAttendees() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionService_GroupPermissions(t *testing.T) {
	service := NewPermissionService()

	groupID := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	nonMemberID := uuid.New()

	userGroups := []GroupMember{
		{GroupID: groupID, UserID: ownerID, Role: GroupRoleOwner, JoinedAt: time.Now()},
		{GroupID: groupID, UserID: adminID, Role: GroupRoleAdmin, JoinedAt: time.Now()},
		{GroupID: groupID, UserID: memberID, Role: GroupRoleMember, JoinedAt: time.Now()},
	}

	t.Run("CanManageGroup", func(t *testing.T) {
		tests := []struct {
			userID uuid.UUID
			want   bool
		}{
			{ownerID, true},
			{adminID, true},
			{memberID, false},
			{nonMemberID, false},
		}

		for _, tt := range tests {
			got := service.CanManageGroup(groupID, tt.userID, userGroups)
			if got != tt.want {
				t.Errorf("CanManageGroup(%v) = %v, want %v", tt.userID, got, tt.want)
			}
		}
	})

	t.Run("CanDeleteGroup", func(t *testing.T) {
		tests := []struct {
			userID uuid.UUID
			want   bool
		}{
			{ownerID, true},
			{adminID, false},
			{memberID, false},
			{nonMemberID, false},
		}

		for _, tt := range tests {
			got := service.CanDeleteGroup(groupID, tt.userID, userGroups)
			if got != tt.want {
				t.Errorf("CanDeleteGroup(%v) = %v, want %v", tt.userID, got, tt.want)
			}
		}
	})
}

func TestPermissionService_CanRemoveMembers(t *testing.T) {
	service := NewPermissionService()

	groupID := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()
	targetMemberID := uuid.New()

	userGroups := []GroupMember{
		{GroupID: groupID, UserID: ownerID, Role: GroupRoleOwner, JoinedAt: time.Now()},
		{GroupID: groupID, UserID: adminID, Role: GroupRoleAdmin, JoinedAt: time.Now()},
		{GroupID: groupID, UserID: memberID, Role: GroupRoleMember, JoinedAt: time.Now()},
		{GroupID: groupID, UserID: targetMemberID, Role: GroupRoleMember, JoinedAt: time.Now()},
	}

	tests := []struct {
		name         string
		userID       uuid.UUID
		targetUserID uuid.UUID
		want         bool
	}{
		{"owner can remove member", ownerID, targetMemberID, true},
		{"owner cannot remove admin", ownerID, adminID, true},          // Owner can remove admin
		{"owner cannot remove another owner", ownerID, ownerID, false}, // Cannot remove self
		{"admin can remove member", adminID, targetMemberID, true},
		{"admin cannot remove owner", adminID, ownerID, false},
		{"member cannot remove anyone", memberID, targetMemberID, false},
		{"cannot remove self", memberID, memberID, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanRemoveMembers(groupID, tt.userID, userGroups, tt.targetUserID)
			if got != tt.want {
				t.Errorf("CanRemoveMembers() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionService_CanUpdateMemberRole(t *testing.T) {
	service := NewPermissionService()

	groupID := uuid.New()
	ownerID := uuid.New()
	adminID := uuid.New()
	memberID := uuid.New()

	userGroups := []GroupMember{
		{GroupID: groupID, UserID: ownerID, Role: GroupRoleOwner, JoinedAt: time.Now()},
		{GroupID: groupID, UserID: adminID, Role: GroupRoleAdmin, JoinedAt: time.Now()},
		{GroupID: groupID, UserID: memberID, Role: GroupRoleMember, JoinedAt: time.Now()},
	}

	tests := []struct {
		name         string
		userID       uuid.UUID
		targetUserID uuid.UUID
		newRole      GroupRole
		want         bool
	}{
		{"owner can promote member to admin", ownerID, memberID, GroupRoleAdmin, true},
		{"owner can demote admin to member", ownerID, adminID, GroupRoleMember, true},
		{"owner cannot promote to owner", ownerID, memberID, GroupRoleOwner, false},
		{"owner cannot update another owner", ownerID, ownerID, GroupRoleAdmin, false}, // Cannot update self
		{"admin cannot update roles", adminID, memberID, GroupRoleAdmin, false},
		{"member cannot update roles", memberID, adminID, GroupRoleMember, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := service.CanUpdateMemberRole(groupID, tt.userID, userGroups, tt.targetUserID, tt.newRole)
			if got != tt.want {
				t.Errorf("CanUpdateMemberRole() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPermissionService_FilterVisibleEvents(t *testing.T) {
	service := NewPermissionService()

	hostID := uuid.New()
	userID := uuid.New()
	groupID := uuid.New()

	userGroups := []GroupMember{
		{GroupID: groupID, UserID: userID, Role: GroupRoleMember, JoinedAt: time.Now()},
	}

	events := []Event{
		{
			ID:         uuid.New(),
			HostUserID: hostID,
			Visibility: EventVisibilityPublic,
		},
		{
			ID:         uuid.New(),
			HostUserID: hostID,
			Visibility: EventVisibilityPrivate,
		},
		{
			ID:         uuid.New(),
			HostUserID: hostID,
			GroupID:    &groupID,
			Visibility: EventVisibilityGroupOnly,
		},
		{
			ID:         uuid.New(),
			HostUserID: userID, // User's own event
			Visibility: EventVisibilityPrivate,
		},
	}

	visibleEvents := service.FilterVisibleEvents(events, userID, userGroups)

	// Should see: public event, group event, and own private event (3 total)
	if len(visibleEvents) != 3 {
		t.Errorf("Expected 3 visible events, got %d", len(visibleEvents))
	}
}

func TestPermissionService_GetUserEventPermissions(t *testing.T) {
	service := NewPermissionService()

	hostID := uuid.New()
	userID := uuid.New()

	event := Event{
		ID:         uuid.New(),
		HostUserID: hostID,
		Visibility: EventVisibilityPublic,
	}

	permissions := service.GetUserEventPermissions(&event, userID, []GroupMember{})

	// Regular user should be able to view and RSVP to public event
	expectedPermissions := []Permission{PermissionViewEvent, PermissionRSVPToEvent, PermissionViewAttendees}

	if len(permissions) != len(expectedPermissions) {
		t.Errorf("Expected %d permissions, got %d", len(expectedPermissions), len(permissions))
	}

	for _, expected := range expectedPermissions {
		found := false
		for _, perm := range permissions {
			if perm == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected permission %v not found", expected)
		}
	}
}
