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

func TestGroupRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewGroupRepository(db)
	ctx := context.Background()

	// Create test user first
	user := createTestUser(t, db)

	group := &domain.Group{
		ID:          uuid.New(),
		Name:        "Test Group",
		OwnerUserID: user.ID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		IsActive:    true,
	}

	err := repo.Create(ctx, group)
	require.NoError(t, err)

	// Verify group was created
	retrieved, err := repo.GetByID(ctx, group.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, group.Name, retrieved.Name)
	assert.Equal(t, group.OwnerUserID, retrieved.OwnerUserID)

	// Verify owner was added as member
	member, err := repo.GetMember(ctx, group.ID, user.ID)
	require.NoError(t, err)
	require.NotNil(t, member)
	assert.Equal(t, domain.GroupRoleOwner, member.Role)
}

func TestGroupRepository_MemberManagement(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewGroupRepository(db)
	ctx := context.Background()

	// Create test users
	owner := createTestUser(t, db)
	member := createTestUser(t, db)

	// Create group
	group := createTestGroup(t, db, owner.ID)

	// Add member
	groupMember := &domain.GroupMember{
		GroupID:  group.ID,
		UserID:   member.ID,
		Role:     domain.GroupRoleMember,
		JoinedAt: time.Now(),
	}

	err := repo.AddMember(ctx, groupMember)
	require.NoError(t, err)

	// Verify member was added
	retrieved, err := repo.GetMember(ctx, group.ID, member.ID)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	assert.Equal(t, domain.GroupRoleMember, retrieved.Role)

	// Check membership
	isMember, err := repo.IsMember(ctx, group.ID, member.ID)
	require.NoError(t, err)
	assert.True(t, isMember)

	// Get member role
	role, err := repo.GetMemberRole(ctx, group.ID, member.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.GroupRoleMember, role)

	// Update member role
	err = repo.UpdateMemberRole(ctx, group.ID, member.ID, domain.GroupRoleAdmin)
	require.NoError(t, err)

	// Verify role update
	updatedRole, err := repo.GetMemberRole(ctx, group.ID, member.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.GroupRoleAdmin, updatedRole)

	// Get all members
	members, err := repo.GetGroupMembers(ctx, group.ID)
	require.NoError(t, err)
	assert.Len(t, members, 2) // Owner + added member

	// Get member count
	count, err := repo.GetMemberCount(ctx, group.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	// Remove member
	err = repo.RemoveMember(ctx, group.ID, member.ID)
	require.NoError(t, err)

	// Verify removal
	isStillMember, err := repo.IsMember(ctx, group.ID, member.ID)
	require.NoError(t, err)
	assert.False(t, isStillMember)
}

func TestGroupRepository_GetUserGroups(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewGroupRepository(db)
	ctx := context.Background()

	// Create test users
	user1 := createTestUser(t, db)
	user2 := createTestUser(t, db)

	// Create groups
	group1 := createTestGroup(t, db, user1.ID)
	group2 := createTestGroup(t, db, user2.ID)

	// Add user1 to group2
	member := &domain.GroupMember{
		GroupID:  group2.ID,
		UserID:   user1.ID,
		Role:     domain.GroupRoleMember,
		JoinedAt: time.Now(),
	}

	err := repo.AddMember(ctx, member)
	require.NoError(t, err)

	// Get user1's groups
	groups, err := repo.GetUserGroups(ctx, user1.ID)
	require.NoError(t, err)
	assert.Len(t, groups, 2) // Owns group1, member of group2

	// Get groups owned by user1
	ownedGroups, err := repo.GetGroupsByOwner(ctx, user1.ID)
	require.NoError(t, err)
	assert.Len(t, ownedGroups, 1)
	assert.Equal(t, group1.ID, ownedGroups[0].ID)
}

func TestGroupRepository_Permissions(t *testing.T) {
	db := setupTestDB(t)
	defer cleanupTestDB(t, db)

	repo := NewGroupRepository(db)
	ctx := context.Background()

	// Create test users
	owner := createTestUser(t, db)
	admin := createTestUser(t, db)
	member := createTestUser(t, db)
	nonMember := createTestUser(t, db)

	// Create group
	group := createTestGroup(t, db, owner.ID)

	// Add admin
	adminMember := &domain.GroupMember{
		GroupID:  group.ID,
		UserID:   admin.ID,
		Role:     domain.GroupRoleAdmin,
		JoinedAt: time.Now(),
	}
	err := repo.AddMember(ctx, adminMember)
	require.NoError(t, err)

	// Add regular member
	regularMember := &domain.GroupMember{
		GroupID:  group.ID,
		UserID:   member.ID,
		Role:     domain.GroupRoleMember,
		JoinedAt: time.Now(),
	}
	err = repo.AddMember(ctx, regularMember)
	require.NoError(t, err)

	// Test owner permissions
	isOwner, err := repo.IsGroupOwner(ctx, group.ID, owner.ID)
	require.NoError(t, err)
	assert.True(t, isOwner)

	canManage, err := repo.CanUserManageGroup(ctx, group.ID, owner.ID)
	require.NoError(t, err)
	assert.True(t, canManage)

	// Test admin permissions
	isOwner, err = repo.IsGroupOwner(ctx, group.ID, admin.ID)
	require.NoError(t, err)
	assert.False(t, isOwner)

	canManage, err = repo.CanUserManageGroup(ctx, group.ID, admin.ID)
	require.NoError(t, err)
	assert.True(t, canManage)

	// Test member permissions
	canManage, err = repo.CanUserManageGroup(ctx, group.ID, member.ID)
	require.NoError(t, err)
	assert.False(t, canManage)

	canAccess, err := repo.CanUserAccessGroup(ctx, group.ID, member.ID)
	require.NoError(t, err)
	assert.True(t, canAccess)

	// Test non-member permissions
	canAccess, err = repo.CanUserAccessGroup(ctx, group.ID, nonMember.ID)
	require.NoError(t, err)
	assert.False(t, canAccess)
}
