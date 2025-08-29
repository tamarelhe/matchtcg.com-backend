package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/repository"
	"github.com/matchtcg/backend/internal/service"
)

type groupRepository struct {
	db service.DB
}

// NewGroupRepository creates a new PostgreSQL group repository
func NewGroupRepository(db service.DB) repository.GroupRepository {
	return &groupRepository{db: db}
}

// Create creates a new group
func (r *groupRepository) Create(ctx context.Context, group *domain.Group) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create the group
	query := `
		INSERT INTO groups (id, name, description, owner_user_id, created_at, updated_at, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err = tx.Exec(ctx, query,
		group.ID,
		group.Name,
		group.Description,
		group.OwnerUserID,
		group.CreatedAt,
		group.UpdatedAt,
		group.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	// Add the owner as a member with owner role
	memberQuery := `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)`

	_, err = tx.Exec(ctx, memberQuery,
		group.ID,
		group.OwnerUserID,
		domain.GroupRoleOwner,
		group.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add owner as member: %w", err)
	}

	return tx.Commit(ctx)
}

// GetByID retrieves a group by ID
func (r *groupRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Group, error) {
	query := `
		SELECT id, name, description, owner_user_id, created_at, updated_at, is_active
		FROM groups
		WHERE id = $1`

	var group domain.Group
	err := r.db.QueryRow(ctx, query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Description,
		&group.OwnerUserID,
		&group.CreatedAt,
		&group.UpdatedAt,
		&group.IsActive,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group by ID: %w", err)
	}

	return &group, nil
}

// GetByIDWithMembers retrieves a group with its members by ID
func (r *groupRepository) GetByIDWithMembers(ctx context.Context, id uuid.UUID) (*domain.GroupWithMembers, error) {
	group, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if group == nil {
		return nil, nil
	}

	members, err := r.GetGroupMembers(ctx, id)
	if err != nil {
		return nil, err
	}

	var membersList []domain.GroupMember
	for _, member := range members {
		membersList = append(membersList, *member)
	}

	return &domain.GroupWithMembers{
		Group:   *group,
		Members: membersList,
	}, nil
}

// Update updates a group
func (r *groupRepository) Update(ctx context.Context, group *domain.Group) error {
	query := `
		UPDATE groups
		SET name = $2, description = $3, owner_user_id = $4, updated_at = $5, is_active = $6
		WHERE id = $1`

	result, err := r.db.Exec(ctx, query,
		group.ID,
		group.Name,
		group.Description,
		group.OwnerUserID,
		group.UpdatedAt,
		group.IsActive,
	)

	if err != nil {
		return fmt.Errorf("failed to update group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// Delete deletes a group
func (r *groupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM groups WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete group: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group not found")
	}

	return nil
}

// GetUserGroups retrieves groups for a specific user
func (r *groupRepository) GetUserGroups(ctx context.Context, userID uuid.UUID) ([]*domain.Group, error) {
	query := `
		SELECT g.id, g.name, g.description, g.owner_user_id, g.created_at, g.updated_at, g.is_active
		FROM groups g
		INNER JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1 AND g.is_active = true
		ORDER BY g.created_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		var group domain.Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.OwnerUserID,
			&group.CreatedAt, &group.UpdatedAt, &group.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, &group)
	}

	return groups, nil
}

// GetUserGroupsWithMembers retrieves groups with members for a specific user
func (r *groupRepository) GetUserGroupsWithMembers(ctx context.Context, userID uuid.UUID) ([]*domain.GroupWithMembers, error) {
	groups, err := r.GetUserGroups(ctx, userID)
	if err != nil {
		return nil, err
	}

	var groupsWithMembers []*domain.GroupWithMembers
	for _, group := range groups {
		groupWithMembers, err := r.GetByIDWithMembers(ctx, group.ID)
		if err != nil {
			return nil, err
		}
		groupsWithMembers = append(groupsWithMembers, groupWithMembers)
	}

	return groupsWithMembers, nil
}

// GetGroupsByOwner retrieves groups owned by a specific user
func (r *groupRepository) GetGroupsByOwner(ctx context.Context, ownerID uuid.UUID) ([]*domain.Group, error) {
	query := `
		SELECT id, name, description, owner_user_id, created_at, updated_at, is_active
		FROM groups
		WHERE owner_user_id = $1 AND is_active = true
		ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get groups by owner: %w", err)
	}
	defer rows.Close()

	var groups []*domain.Group
	for rows.Next() {
		var group domain.Group
		if err := rows.Scan(&group.ID, &group.Name, &group.Description, &group.OwnerUserID,
			&group.CreatedAt, &group.UpdatedAt, &group.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, &group)
	}

	return groups, nil
}

// AddMember adds a member to a group
func (r *groupRepository) AddMember(ctx context.Context, member *domain.GroupMember) error {
	query := `
		INSERT INTO group_members (group_id, user_id, role, joined_at)
		VALUES ($1, $2, $3, $4)`

	_, err := r.db.Exec(ctx, query,
		member.GroupID,
		member.UserID,
		member.Role,
		member.JoinedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to add group member: %w", err)
	}

	return nil
}

// GetMember retrieves a specific group member
func (r *groupRepository) GetMember(ctx context.Context, groupID, userID uuid.UUID) (*domain.GroupMember, error) {
	query := `
		SELECT group_id, user_id, role, joined_at
		FROM group_members
		WHERE group_id = $1 AND user_id = $2`

	var member domain.GroupMember
	err := r.db.QueryRow(ctx, query, groupID, userID).Scan(
		&member.GroupID,
		&member.UserID,
		&member.Role,
		&member.JoinedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group member: %w", err)
	}

	return &member, nil
}

// UpdateMemberRole updates a member's role in a group
func (r *groupRepository) UpdateMemberRole(ctx context.Context, groupID, userID uuid.UUID, role domain.GroupRole) error {
	query := `
		UPDATE group_members
		SET role = $3
		WHERE group_id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, groupID, userID, role)
	if err != nil {
		return fmt.Errorf("failed to update member role: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group member not found")
	}

	return nil
}

// RemoveMember removes a member from a group
func (r *groupRepository) RemoveMember(ctx context.Context, groupID, userID uuid.UUID) error {
	query := `DELETE FROM group_members WHERE group_id = $1 AND user_id = $2`

	result, err := r.db.Exec(ctx, query, groupID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove group member: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("group member not found")
	}

	return nil
}

// GetGroupMembers retrieves all members of a group
func (r *groupRepository) GetGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*domain.GroupMember, error) {
	query := `
		SELECT group_id, user_id, role, joined_at
		FROM group_members
		WHERE group_id = $1
		ORDER BY joined_at ASC`

	rows, err := r.db.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	var members []*domain.GroupMember
	for rows.Next() {
		var member domain.GroupMember
		if err := rows.Scan(&member.GroupID, &member.UserID, &member.Role, &member.JoinedAt); err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		members = append(members, &member)
	}

	return members, nil
}

// GetMembersByRole retrieves members of a group with a specific role
func (r *groupRepository) GetMembersByRole(ctx context.Context, groupID uuid.UUID, role domain.GroupRole) ([]*domain.GroupMember, error) {
	query := `
		SELECT group_id, user_id, role, joined_at
		FROM group_members
		WHERE group_id = $1 AND role = $2
		ORDER BY joined_at ASC`

	rows, err := r.db.Query(ctx, query, groupID, role)
	if err != nil {
		return nil, fmt.Errorf("failed to get members by role: %w", err)
	}
	defer rows.Close()

	var members []*domain.GroupMember
	for rows.Next() {
		var member domain.GroupMember
		if err := rows.Scan(&member.GroupID, &member.UserID, &member.Role, &member.JoinedAt); err != nil {
			return nil, fmt.Errorf("failed to scan group member: %w", err)
		}
		members = append(members, &member)
	}

	return members, nil
}

// IsMember checks if a user is a member of a group
func (r *groupRepository) IsMember(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM group_members WHERE group_id = $1 AND user_id = $2)`

	var exists bool
	err := r.db.QueryRow(ctx, query, groupID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check group membership: %w", err)
	}

	return exists, nil
}

// GetMemberRole retrieves a user's role in a group
func (r *groupRepository) GetMemberRole(ctx context.Context, groupID, userID uuid.UUID) (domain.GroupRole, error) {
	query := `SELECT role FROM group_members WHERE group_id = $1 AND user_id = $2`

	var role domain.GroupRole
	err := r.db.QueryRow(ctx, query, groupID, userID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("user is not a member of the group")
		}
		return "", fmt.Errorf("failed to get member role: %w", err)
	}

	return role, nil
}

// CanUserAccessGroup checks if a user can access a group
func (r *groupRepository) CanUserAccessGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	// For now, any member can access the group
	return r.IsMember(ctx, groupID, userID)
}

// CanUserManageGroup checks if a user can manage a group (admin or owner)
func (r *groupRepository) CanUserManageGroup(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	role, err := r.GetMemberRole(ctx, groupID, userID)
	if err != nil {
		return false, err
	}

	return role == domain.GroupRoleOwner || role == domain.GroupRoleAdmin, nil
}

// IsGroupOwner checks if a user is the owner of a group
func (r *groupRepository) IsGroupOwner(ctx context.Context, groupID, userID uuid.UUID) (bool, error) {
	query := `SELECT owner_user_id FROM groups WHERE id = $1`

	var ownerID uuid.UUID
	err := r.db.QueryRow(ctx, query, groupID).Scan(&ownerID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, fmt.Errorf("group not found")
		}
		return false, fmt.Errorf("failed to get group owner: %w", err)
	}

	return ownerID == userID, nil
}

// GetMemberCount retrieves the number of members in a group
func (r *groupRepository) GetMemberCount(ctx context.Context, groupID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM group_members WHERE group_id = $1`

	var count int
	err := r.db.QueryRow(ctx, query, groupID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get member count: %w", err)
	}

	return count, nil
}
