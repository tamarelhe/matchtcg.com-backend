package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockGroupRepository is a mock implementation of GroupRepository
type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) Create(ctx context.Context, group *domain.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*domain.GroupWithMembers, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GroupWithMembers), args.Error(1)
}

func (m *MockGroupRepository) Update(ctx context.Context, group *domain.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGroupRepository) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*domain.Group, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Group), args.Error(1)
}

func (m *MockGroupRepository) GetUserGroupsWithMembers(ctx context.Context, userID uuid.UUID) ([]*domain.GroupWithMembers, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GroupWithMembers), args.Error(1)
}

func (m *MockGroupRepository) GetGroupsByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Group, error) {
	args := m.Called(ctx, ownerID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Group), args.Error(1)
}

func (m *MockGroupRepository) AddMember(ctx context.Context, member *domain.GroupMember) error {
	args := m.Called(ctx, member)
	return args.Error(0)
}

func (m *MockGroupRepository) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error) {
	args := m.Called(ctx, groupID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.GroupMember), args.Error(1)
}

func (m *MockGroupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role domain.GroupRole) error {
	args := m.Called(ctx, groupID, userID, role)
	return args.Error(0)
}

func (m *MockGroupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	args := m.Called(ctx, groupID, userID)
	return args.Error(0)
}

func (m *MockGroupRepository) GetGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMember, error) {
	args := m.Called(ctx, groupID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GroupMember), args.Error(1)
}

func (m *MockGroupRepository) GetMembersByRole(ctx context.Context, groupID uuid.UUID, role domain.GroupRole) ([]*domain.GroupMember, error) {
	args := m.Called(ctx, groupID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.GroupMember), args.Error(1)
}

func (m *MockGroupRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupRepository) GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (domain.GroupRole, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Get(0).(domain.GroupRole), args.Error(1)
}

func (m *MockGroupRepository) CanUserAccessGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupRepository) CanUserManageGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupRepository) IsGroupOwner(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	args := m.Called(ctx, groupID, userID)
	return args.Bool(0), args.Error(1)
}

func (m *MockGroupRepository) GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	args := m.Called(ctx, groupID)
	return args.Int(0), args.Error(1)
}

func TestCreateGroupUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful group creation", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockUserRepo := new(MockUserRepository)
		useCase := NewCreateGroupUseCase(mockGroupRepo, mockUserRepo)

		ownerID := uuid.New()
		owner := &domain.User{
			ID:       ownerID,
			Email:    "owner@example.com",
			IsActive: true,
		}

		req := &CreateGroupRequest{
			Name:        "Test Group",
			Description: stringPtr("A test group"),
			OwnerUserID: ownerID,
		}

		mockUserRepo.On("GetByID", ctx, ownerID).Return(owner, nil)
		mockGroupRepo.On("Create", ctx, mock.AnythingOfType("*domain.Group")).Return(nil)

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "Test Group", resp.Group.Name)
		assert.Equal(t, ownerID, resp.Group.OwnerUserID)
		assert.True(t, resp.Group.IsActive)
		mockUserRepo.AssertExpectations(t)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("invalid owner user", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockUserRepo := new(MockUserRepository)
		useCase := NewCreateGroupUseCase(mockGroupRepo, mockUserRepo)

		ownerID := uuid.New()
		req := &CreateGroupRequest{
			Name:        "Test Group",
			OwnerUserID: ownerID,
		}

		mockUserRepo.On("GetByID", ctx, ownerID).Return(nil, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrInvalidGroupOwner, err)
		assert.Nil(t, resp)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("empty group name", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockUserRepo := new(MockUserRepository)
		useCase := NewCreateGroupUseCase(mockGroupRepo, mockUserRepo)

		ownerID := uuid.New()
		owner := &domain.User{
			ID:       ownerID,
			Email:    "owner@example.com",
			IsActive: true,
		}

		req := &CreateGroupRequest{
			Name:        "",
			OwnerUserID: ownerID,
		}

		mockUserRepo.On("GetByID", ctx, ownerID).Return(owner, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrEmptyGroupName, err)
		assert.Nil(t, resp)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUpdateGroupUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful group update", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewUpdateGroupUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		existingGroup := &domain.Group{
			ID:          groupID,
			Name:        "Old Name",
			Description: stringPtr("Old description"),
			OwnerUserID: userID,
			IsActive:    true,
		}

		req := &UpdateGroupRequest{
			GroupID: groupID,
			Name:    stringPtr("New Name"),
			UserID:  userID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(existingGroup, nil)
		mockGroupRepo.On("CanUserManageGroup", ctx, groupID, userID).Return(true, nil)
		mockGroupRepo.On("Update", ctx, mock.AnythingOfType("*domain.Group")).Return(nil)

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "New Name", resp.Group.Name)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewUpdateGroupUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()

		req := &UpdateGroupRequest{
			GroupID: groupID,
			Name:    stringPtr("New Name"),
			UserID:  userID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(nil, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrGroupNotFound, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewUpdateGroupUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		existingGroup := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: uuid.New(), // Different owner
			IsActive:    true,
		}

		req := &UpdateGroupRequest{
			GroupID: groupID,
			Name:    stringPtr("New Name"),
			UserID:  userID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(existingGroup, nil)
		mockGroupRepo.On("CanUserManageGroup", ctx, groupID, userID).Return(false, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedGroupAccess, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})
}

func TestDeleteGroupUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful group deletion", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewDeleteGroupUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		existingGroup := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: userID,
			IsActive:    true,
		}

		req := &DeleteGroupRequest{
			GroupID: groupID,
			UserID:  userID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(existingGroup, nil)
		mockGroupRepo.On("IsGroupOwner", ctx, groupID, userID).Return(true, nil)
		mockGroupRepo.On("Delete", ctx, groupID).Return(nil)

		err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewDeleteGroupUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()

		req := &DeleteGroupRequest{
			GroupID: groupID,
			UserID:  userID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(nil, nil)

		err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrGroupNotFound, err)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("non-owner cannot delete", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewDeleteGroupUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		existingGroup := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: uuid.New(), // Different owner
			IsActive:    true,
		}

		req := &DeleteGroupRequest{
			GroupID: groupID,
			UserID:  userID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(existingGroup, nil)
		mockGroupRepo.On("IsGroupOwner", ctx, groupID, userID).Return(false, nil)

		err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrCannotDeleteGroup, err)
		mockGroupRepo.AssertExpectations(t)
	})
}

func TestInviteGroupMemberUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful member invitation", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockUserRepo := new(MockUserRepository)
		useCase := NewInviteGroupMemberUseCase(mockGroupRepo, mockUserRepo)

		groupID := uuid.New()
		userID := uuid.New()
		inviterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: inviterID,
			IsActive:    true,
		}

		user := &domain.User{
			ID:       userID,
			Email:    "user@example.com",
			IsActive: true,
		}

		req := &InviteGroupMemberRequest{
			GroupID:   groupID,
			UserID:    userID,
			Role:      domain.GroupRoleMember,
			InviterID: inviterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockGroupRepo.On("CanUserManageGroup", ctx, groupID, inviterID).Return(true, nil)
		mockGroupRepo.On("IsMember", ctx, groupID, userID).Return(false, nil)
		mockGroupRepo.On("GetMemberRole", ctx, groupID, inviterID).Return(domain.GroupRoleOwner, nil)
		mockGroupRepo.On("AddMember", ctx, mock.AnythingOfType("*domain.GroupMember")).Return(nil)

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, userID, resp.Member.UserID)
		assert.Equal(t, domain.GroupRoleMember, resp.Member.Role)
		mockGroupRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("user already member", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockUserRepo := new(MockUserRepository)
		useCase := NewInviteGroupMemberUseCase(mockGroupRepo, mockUserRepo)

		groupID := uuid.New()
		userID := uuid.New()
		inviterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: inviterID,
			IsActive:    true,
		}

		user := &domain.User{
			ID:       userID,
			Email:    "user@example.com",
			IsActive: true,
		}

		req := &InviteGroupMemberRequest{
			GroupID:   groupID,
			UserID:    userID,
			Role:      domain.GroupRoleMember,
			InviterID: inviterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockGroupRepo.On("CanUserManageGroup", ctx, groupID, inviterID).Return(true, nil)
		mockGroupRepo.On("IsMember", ctx, groupID, userID).Return(true, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "already a member")
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockUserRepo := new(MockUserRepository)
		useCase := NewInviteGroupMemberUseCase(mockGroupRepo, mockUserRepo)

		groupID := uuid.New()
		userID := uuid.New()
		inviterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: uuid.New(), // Different owner
			IsActive:    true,
		}

		user := &domain.User{
			ID:       userID,
			Email:    "user@example.com",
			IsActive: true,
		}

		req := &InviteGroupMemberRequest{
			GroupID:   groupID,
			UserID:    userID,
			Role:      domain.GroupRoleMember,
			InviterID: inviterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockUserRepo.On("GetByID", ctx, userID).Return(user, nil)
		mockGroupRepo.On("CanUserManageGroup", ctx, groupID, inviterID).Return(false, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrInsufficientPermissions, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestRemoveGroupMemberUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful member removal", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewRemoveGroupMemberUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		removerID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: removerID,
			IsActive:    true,
		}

		member := &domain.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    domain.GroupRoleMember,
		}

		req := &RemoveGroupMemberRequest{
			GroupID:   groupID,
			UserID:    userID,
			RemoverID: removerID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("GetMember", ctx, groupID, userID).Return(member, nil)
		mockGroupRepo.On("CanUserManageGroup", ctx, groupID, removerID).Return(true, nil)
		mockGroupRepo.On("GetMemberRole", ctx, groupID, removerID).Return(domain.GroupRoleOwner, nil)
		mockGroupRepo.On("GetMemberRole", ctx, groupID, userID).Return(domain.GroupRoleMember, nil)
		mockGroupRepo.On("RemoveMember", ctx, groupID, userID).Return(nil)

		err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("user removes themselves", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewRemoveGroupMemberUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: uuid.New(), // Different owner
			IsActive:    true,
		}

		member := &domain.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    domain.GroupRoleMember,
		}

		req := &RemoveGroupMemberRequest{
			GroupID:   groupID,
			UserID:    userID,
			RemoverID: userID, // Same user removing themselves
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("GetMember", ctx, groupID, userID).Return(member, nil)
		// No need for permission checks when removing self
		mockGroupRepo.On("RemoveMember", ctx, groupID, userID).Return(nil)

		err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("cannot remove owner", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewRemoveGroupMemberUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		removerID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: userID,
			IsActive:    true,
		}

		member := &domain.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    domain.GroupRoleOwner,
		}

		req := &RemoveGroupMemberRequest{
			GroupID:   groupID,
			UserID:    userID,
			RemoverID: removerID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("GetMember", ctx, groupID, userID).Return(member, nil)

		err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrCannotRemoveOwner, err)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("user not member", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewRemoveGroupMemberUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		removerID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: removerID,
			IsActive:    true,
		}

		req := &RemoveGroupMemberRequest{
			GroupID:   groupID,
			UserID:    userID,
			RemoverID: removerID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("GetMember", ctx, groupID, userID).Return(nil, nil)

		err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrUserNotGroupMember, err)
		mockGroupRepo.AssertExpectations(t)
	})
}

func TestUpdateMemberRoleUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful role update", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewUpdateMemberRoleUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		updaterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: updaterID,
			IsActive:    true,
		}

		member := &domain.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    domain.GroupRoleMember,
		}

		updatedMember := &domain.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    domain.GroupRoleAdmin,
		}

		req := &UpdateMemberRoleRequest{
			GroupID:   groupID,
			UserID:    userID,
			NewRole:   domain.GroupRoleAdmin,
			UpdaterID: updaterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("GetMember", ctx, groupID, userID).Return(member, nil).Once()
		mockGroupRepo.On("GetMemberRole", ctx, groupID, updaterID).Return(domain.GroupRoleOwner, nil)
		mockGroupRepo.On("UpdateMemberRole", ctx, groupID, userID, domain.GroupRoleAdmin).Return(nil)
		mockGroupRepo.On("GetMember", ctx, groupID, userID).Return(updatedMember, nil).Once()

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, domain.GroupRoleAdmin, resp.Member.Role)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("cannot update owner role", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewUpdateMemberRoleUseCase(mockGroupRepo)

		groupID := uuid.New()
		//userID := uuid.New()
		updaterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: updaterID,
			IsActive:    true,
		}

		member := &domain.GroupMember{
			GroupID: groupID,
			UserID:  updaterID,
			Role:    domain.GroupRoleOwner,
		}

		req := &UpdateMemberRoleRequest{
			GroupID:   groupID,
			UserID:    updaterID,
			NewRole:   domain.GroupRoleAdmin,
			UpdaterID: updaterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("GetMember", ctx, groupID, updaterID).Return(member, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrInsufficientPermissions, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("insufficient permissions", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewUpdateMemberRoleUseCase(mockGroupRepo)

		groupID := uuid.New()
		userID := uuid.New()
		updaterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: uuid.New(), // Different owner
			IsActive:    true,
		}

		member := &domain.GroupMember{
			GroupID: groupID,
			UserID:  userID,
			Role:    domain.GroupRoleMember,
		}

		req := &UpdateMemberRoleRequest{
			GroupID:   groupID,
			UserID:    userID,
			NewRole:   domain.GroupRoleAdmin,
			UpdaterID: updaterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("GetMember", ctx, groupID, userID).Return(member, nil)
		mockGroupRepo.On("GetMemberRole", ctx, groupID, updaterID).Return(domain.GroupRoleMember, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrInsufficientPermissions, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})
}

func TestGetGroupMembersUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful members retrieval", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewGetGroupMembersUseCase(mockGroupRepo)

		groupID := uuid.New()
		requesterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: requesterID,
			IsActive:    true,
		}

		members := []*domain.GroupMember{
			{
				GroupID: groupID,
				UserID:  requesterID,
				Role:    domain.GroupRoleOwner,
			},
			{
				GroupID: groupID,
				UserID:  uuid.New(),
				Role:    domain.GroupRoleMember,
			},
		}

		req := &GetGroupMembersRequest{
			GroupID:     groupID,
			RequesterID: requesterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("CanUserAccessGroup", ctx, groupID, requesterID).Return(true, nil)
		mockGroupRepo.On("GetGroupMembers", ctx, groupID).Return(members, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Members, 2)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewGetGroupMembersUseCase(mockGroupRepo)

		groupID := uuid.New()
		requesterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: uuid.New(), // Different owner
			IsActive:    true,
		}

		req := &GetGroupMembersRequest{
			GroupID:     groupID,
			RequesterID: requesterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("CanUserAccessGroup", ctx, groupID, requesterID).Return(false, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedGroupAccess, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		useCase := NewGetGroupMembersUseCase(mockGroupRepo)

		groupID := uuid.New()
		requesterID := uuid.New()

		req := &GetGroupMembersRequest{
			GroupID:     groupID,
			RequesterID: requesterID,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(nil, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrGroupNotFound, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})
}

func TestGetGroupEventsUseCase_Execute(t *testing.T) {
	ctx := context.Background()

	t.Run("successful group events retrieval", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockEventRepo := new(MockEventRepository)
		useCase := NewGetGroupEventsUseCase(mockGroupRepo, mockEventRepo)

		groupID := uuid.New()
		requesterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: requesterID,
			IsActive:    true,
		}

		events := []*domain.Event{
			{
				ID:         uuid.New(),
				Title:      "Group Event 1",
				GroupID:    &groupID,
				Visibility: domain.EventVisibilityGroupOnly,
			},
			{
				ID:         uuid.New(),
				Title:      "Group Event 2",
				GroupID:    &groupID,
				Visibility: domain.EventVisibilityPublic,
			},
		}

		req := &GetGroupEventsRequest{
			GroupID:     groupID,
			RequesterID: requesterID,
			Limit:       10,
			Offset:      0,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("CanUserAccessGroup", ctx, groupID, requesterID).Return(true, nil)
		mockEventRepo.On("GetGroupEvents", ctx, groupID, 10, 0).Return(events, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Len(t, resp.Events, 2)
		assert.Equal(t, "Group Event 1", resp.Events[0].Title)
		mockGroupRepo.AssertExpectations(t)
		mockEventRepo.AssertExpectations(t)
	})

	t.Run("unauthorized access", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockEventRepo := new(MockEventRepository)
		useCase := NewGetGroupEventsUseCase(mockGroupRepo, mockEventRepo)

		groupID := uuid.New()
		requesterID := uuid.New()

		group := &domain.Group{
			ID:          groupID,
			Name:        "Test Group",
			OwnerUserID: uuid.New(), // Different owner
			IsActive:    true,
		}

		req := &GetGroupEventsRequest{
			GroupID:     groupID,
			RequesterID: requesterID,
			Limit:       10,
			Offset:      0,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(group, nil)
		mockGroupRepo.On("CanUserAccessGroup", ctx, groupID, requesterID).Return(false, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrUnauthorizedGroupAccess, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})

	t.Run("group not found", func(t *testing.T) {
		mockGroupRepo := new(MockGroupRepository)
		mockEventRepo := new(MockEventRepository)
		useCase := NewGetGroupEventsUseCase(mockGroupRepo, mockEventRepo)

		groupID := uuid.New()
		requesterID := uuid.New()

		req := &GetGroupEventsRequest{
			GroupID:     groupID,
			RequesterID: requesterID,
			Limit:       10,
			Offset:      0,
		}

		mockGroupRepo.On("GetByID", ctx, groupID).Return(nil, nil)

		resp, err := useCase.Execute(ctx, req)

		assert.Error(t, err)
		assert.Equal(t, ErrGroupNotFound, err)
		assert.Nil(t, resp)
		mockGroupRepo.AssertExpectations(t)
	})
}
