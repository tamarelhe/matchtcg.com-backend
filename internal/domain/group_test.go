package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestGroup_Validate(t *testing.T) {
	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}
	longNameStr := string(longName)

	longDescription := make([]byte, 1001)
	for i := range longDescription {
		longDescription[i] = 'a'
	}
	longDescriptionStr := string(longDescription)

	tests := []struct {
		name    string
		group   Group
		wantErr error
	}{
		{
			name: "valid group",
			group: Group{
				ID:          uuid.New(),
				Name:        "Magic Players",
				OwnerUserID: uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				IsActive:    true,
			},
			wantErr: nil,
		},
		{
			name: "valid group with description",
			group: Group{
				ID:          uuid.New(),
				Name:        "Magic Players",
				Description: stringPtr("A group for Magic: The Gathering players"),
				OwnerUserID: uuid.New(),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
				IsActive:    true,
			},
			wantErr: nil,
		},
		{
			name: "empty name",
			group: Group{
				ID:          uuid.New(),
				Name:        "",
				OwnerUserID: uuid.New(),
			},
			wantErr: ErrEmptyGroupName,
		},
		{
			name: "whitespace only name",
			group: Group{
				ID:          uuid.New(),
				Name:        "   ",
				OwnerUserID: uuid.New(),
			},
			wantErr: ErrEmptyGroupName,
		},
		{
			name: "name too long",
			group: Group{
				ID:          uuid.New(),
				Name:        longNameStr,
				OwnerUserID: uuid.New(),
			},
			wantErr: ErrGroupNameTooLong,
		},
		{
			name: "description too long",
			group: Group{
				ID:          uuid.New(),
				Name:        "Valid Name",
				Description: &longDescriptionStr,
				OwnerUserID: uuid.New(),
			},
			wantErr: ErrGroupDescriptionTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.group.Validate()
			if err != tt.wantErr {
				t.Errorf("Group.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGroupMember_Validate(t *testing.T) {
	tests := []struct {
		name    string
		member  GroupMember
		wantErr error
	}{
		{
			name: "valid member",
			member: GroupMember{
				GroupID:  uuid.New(),
				UserID:   uuid.New(),
				Role:     GroupRoleMember,
				JoinedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid admin",
			member: GroupMember{
				GroupID:  uuid.New(),
				UserID:   uuid.New(),
				Role:     GroupRoleAdmin,
				JoinedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "valid owner",
			member: GroupMember{
				GroupID:  uuid.New(),
				UserID:   uuid.New(),
				Role:     GroupRoleOwner,
				JoinedAt: time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "invalid role",
			member: GroupMember{
				GroupID:  uuid.New(),
				UserID:   uuid.New(),
				Role:     GroupRole("invalid"),
				JoinedAt: time.Now(),
			},
			wantErr: ErrInvalidGroupRole,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.member.Validate()
			if err != tt.wantErr {
				t.Errorf("GroupMember.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGroupMember_RoleChecks(t *testing.T) {
	owner := GroupMember{
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
		Role:     GroupRoleOwner,
		JoinedAt: time.Now(),
	}

	admin := GroupMember{
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
		Role:     GroupRoleAdmin,
		JoinedAt: time.Now(),
	}

	member := GroupMember{
		GroupID:  uuid.New(),
		UserID:   uuid.New(),
		Role:     GroupRoleMember,
		JoinedAt: time.Now(),
	}

	t.Run("IsOwner", func(t *testing.T) {
		if !owner.IsOwner() {
			t.Error("Owner should be owner")
		}
		if admin.IsOwner() {
			t.Error("Admin should not be owner")
		}
		if member.IsOwner() {
			t.Error("Member should not be owner")
		}
	})

	t.Run("IsAdmin", func(t *testing.T) {
		if !owner.IsAdmin() {
			t.Error("Owner should be admin")
		}
		if !admin.IsAdmin() {
			t.Error("Admin should be admin")
		}
		if member.IsAdmin() {
			t.Error("Member should not be admin")
		}
	})

	t.Run("CanManageMembers", func(t *testing.T) {
		if !owner.CanManageMembers() {
			t.Error("Owner should be able to manage members")
		}
		if !admin.CanManageMembers() {
			t.Error("Admin should be able to manage members")
		}
		if member.CanManageMembers() {
			t.Error("Member should not be able to manage members")
		}
	})

	t.Run("CanEditGroup", func(t *testing.T) {
		if !owner.CanEditGroup() {
			t.Error("Owner should be able to edit group")
		}
		if !admin.CanEditGroup() {
			t.Error("Admin should be able to edit group")
		}
		if member.CanEditGroup() {
			t.Error("Member should not be able to edit group")
		}
	})

	t.Run("CanDeleteGroup", func(t *testing.T) {
		if !owner.CanDeleteGroup() {
			t.Error("Owner should be able to delete group")
		}
		if admin.CanDeleteGroup() {
			t.Error("Admin should not be able to delete group")
		}
		if member.CanDeleteGroup() {
			t.Error("Member should not be able to delete group")
		}
	})

	t.Run("HasPermission", func(t *testing.T) {
		// Test member permissions
		if !member.HasPermission(GroupRoleMember) {
			t.Error("Member should have member permissions")
		}
		if member.HasPermission(GroupRoleAdmin) {
			t.Error("Member should not have admin permissions")
		}
		if member.HasPermission(GroupRoleOwner) {
			t.Error("Member should not have owner permissions")
		}

		// Test admin permissions
		if !admin.HasPermission(GroupRoleMember) {
			t.Error("Admin should have member permissions")
		}
		if !admin.HasPermission(GroupRoleAdmin) {
			t.Error("Admin should have admin permissions")
		}
		if admin.HasPermission(GroupRoleOwner) {
			t.Error("Admin should not have owner permissions")
		}

		// Test owner permissions
		if !owner.HasPermission(GroupRoleMember) {
			t.Error("Owner should have member permissions")
		}
		if !owner.HasPermission(GroupRoleAdmin) {
			t.Error("Owner should have admin permissions")
		}
		if !owner.HasPermission(GroupRoleOwner) {
			t.Error("Owner should have owner permissions")
		}
	})
}

// Helper function for string pointers
func stringPtr(s string) *string {
	return &s
}
