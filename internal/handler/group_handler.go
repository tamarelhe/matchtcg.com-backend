package handler

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	"github.com/matchtcg/backend/internal/domain"
	"github.com/matchtcg/backend/internal/middleware"
	"github.com/matchtcg/backend/internal/usecase"
)

// GroupHandler handles group management HTTP requests
type GroupHandler struct {
	groupManagementUseCase *usecase.GroupManagementUseCase
}

// CreateGroupRequest represents the group creation request payload
type CreateGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	Description string `json:"description,omitempty" validate:"max=1000"`
}

// UpdateGroupRequest represents the group update request payload
type UpdateGroupRequest struct {
	Name        *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=1000"`
}

// AddMemberRequest represents the add member request payload
type AddMemberRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
	Role   string `json:"role" validate:"required,group_role"`
}

// UpdateMemberRoleRequest represents the update member role request payload
type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,group_role"`
}

// GroupResponse represents the group response
type GroupResponse struct {
	ID          string                `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Owner       *UserInfo             `json:"owner"`
	MemberCount int                   `json:"member_count"`
	UserRole    string                `json:"user_role,omitempty"`
	Members     []GroupMemberResponse `json:"members,omitempty"`
	CreatedAt   string                `json:"created_at"`
	UpdatedAt   string                `json:"updated_at"`
}

// GroupMemberResponse represents a group member
type GroupMemberResponse struct {
	User     UserInfo `json:"user"`
	Role     string   `json:"role"`
	JoinedAt string   `json:"joined_at"`
}

// GroupListResponse represents a paginated list of groups
type GroupListResponse struct {
	Groups []GroupResponse `json:"groups"`
	Total  int             `json:"total"`
}

// NewGroupHandler creates a new group handler
func NewGroupHandler(groupManagementUseCase *usecase.GroupManagementUseCase) *GroupHandler {
	return &GroupHandler{
		groupManagementUseCase: groupManagementUseCase,
	}
}

// stringToGroupRole converts a string to domain.GroupRole
func stringToGroupRole(s string) domain.GroupRole {
	switch s {
	case "member":
		return domain.GroupRoleMember
	case "admin":
		return domain.GroupRoleAdmin
	case "owner":
		return domain.GroupRoleOwner
	default:
		return domain.GroupRoleMember // Default fallback
	}
}

// CreateGroup handles POST /groups
func (h *GroupHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Group name is required")
		return
	}

	// Create group use case request
	createReq := &usecase.CreateGroupRequest{
		OwnerUserID: userUUID,
		Name:        req.Name,
		Description: &req.Description,
	}

	// Execute group creation
	result, err := h.groupManagementUseCase.CreateGroup(r.Context(), createReq)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "group_creation_failed", "Failed to create group")
		return
	}

	// Convert to response format
	response := h.convertToGroupResponse(result, &userUUID, false)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetGroup handles GET /groups/{id}
func (h *GroupHandler) GetGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID")
		return
	}

	// Get requesting user ID (optional)
	var requestingUserID *uuid.UUID
	if userID, ok := middleware.GetUserID(r); ok {
		if parsed, err := uuid.Parse(userID); err == nil {
			requestingUserID = &parsed
		}
	}

	// Check if members should be included
	includeMembers := r.URL.Query().Get("include_members") == "true"

	// Get group with permission checks
	req := &usecase.GetGroupRequest{
		GroupID:          groupID,
		RequestingUserID: requestingUserID,
		IncludeMembers:   includeMembers,
	}

	result, err := h.groupManagementUseCase.GetGroup(r.Context(), req)
	if err != nil {
		switch err {
		case usecase.ErrGroupNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "group_not_found", "Group not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Access denied to this group")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "group_fetch_failed", "Failed to fetch group")
		}
		return
	}

	// Convert to response format
	response := h.convertToGroupResponse(result, requestingUserID, includeMembers)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateGroup handles PUT /groups/{id}
func (h *GroupHandler) UpdateGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID")
		return
	}

	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	var req UpdateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Create update group request
	updateReq := &usecase.UpdateGroupRequest{
		GroupID:     groupID,
		UserID:      userUUID,
		Name:        req.Name,
		Description: req.Description,
	}

	// Execute group update
	result, err := h.groupManagementUseCase.UpdateGroup(r.Context(), updateReq)
	if err != nil {
		switch err {
		case usecase.ErrGroupNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "group_not_found", "Group not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Only group owner or admin can update this group")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "group_update_failed", "Failed to update group")
		}
		return
	}

	// Convert to response format
	response := h.convertToGroupResponse(result, &userUUID, false)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteGroup handles DELETE /groups/{id}
func (h *GroupHandler) DeleteGroup(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID")
		return
	}

	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Create delete group request
	deleteReq := &usecase.DeleteGroupRequest{
		GroupID: groupID,
		UserID:  userUUID,
	}

	// Execute group deletion
	err = h.groupManagementUseCase.DeleteGroup(r.Context(), deleteReq)
	if err != nil {
		switch err {
		case usecase.ErrGroupNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "group_not_found", "Group not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Only group owner can delete this group")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "group_deletion_failed", "Failed to delete group")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Group successfully deleted",
	})
}

// AddGroupMember handles POST /groups/{id}/members
func (h *GroupHandler) AddGroupMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["id"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID")
		return
	}

	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	requestingUserUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	var req AddMemberRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	// Validate required fields
	if req.UserID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "User ID is required")
		return
	}
	if req.Role == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Role is required")
		return
	}

	targetUserUUID, err := uuid.Parse(req.UserID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_target_user_id", "Invalid target user ID")
		return
	}

	// Create add member request
	addMemberReq := &usecase.InviteGroupMemberRequest{
		GroupID:   groupID,
		UserID:    targetUserUUID,
		Role:      stringToGroupRole(req.Role),
		InviterID: requestingUserUUID,
	}

	// Execute add member
	_, err = h.groupManagementUseCase.InviteGroupMember(r.Context(), addMemberReq)
	if err != nil {
		switch err {
		case usecase.ErrGroupNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "group_not_found", "Group not found")
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Only group owner or admin can add members")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "add_member_failed", "Failed to add group member")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member successfully added to group",
	})
}

// RemoveGroupMember handles DELETE /groups/{id}/members/{userId}
func (h *GroupHandler) RemoveGroupMember(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["id"]
	targetUserIDStr := vars["userId"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID")
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	requestingUserUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_requesting_user_id", "Invalid requesting user ID")
		return
	}

	// Create remove member request
	removeMemberReq := &usecase.RemoveGroupMemberRequest{
		GroupID:   groupID,
		RemoverID: requestingUserUUID,
		UserID:    targetUserID,
	}

	// Execute remove member
	err = h.groupManagementUseCase.RemoveGroupMember(r.Context(), removeMemberReq)
	if err != nil {
		switch err {
		case usecase.ErrGroupNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "group_not_found", "Group not found")
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Insufficient permissions to remove member")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "remove_member_failed", "Failed to remove group member")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member successfully removed from group",
	})
}

// UpdateMemberRole handles PUT /groups/{id}/members/{userId}
func (h *GroupHandler) UpdateMemberRole(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	groupIDStr := vars["id"]
	targetUserIDStr := vars["userId"]

	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_group_id", "Invalid group ID")
		return
	}

	targetUserID, err := uuid.Parse(targetUserIDStr)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	requestingUserUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_requesting_user_id", "Invalid requesting user ID")
		return
	}

	var req UpdateMemberRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_request", "Invalid request body")
		return
	}

	if req.Role == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "validation_error", "Role is required")
		return
	}

	// Create update member role request
	updateRoleReq := &usecase.UpdateMemberRoleRequest{
		GroupID:   groupID,
		UpdaterID: requestingUserUUID,
		UserID:    targetUserID,
		NewRole:   stringToGroupRole(req.Role),
	}

	// Execute update member role
	_, err = h.groupManagementUseCase.UpdateMemberRole(r.Context(), updateRoleReq)
	if err != nil {
		switch err {
		case usecase.ErrGroupNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "group_not_found", "Group not found")
		case usecase.ErrUserNotFound:
			h.writeErrorResponse(w, http.StatusNotFound, "user_not_found", "User not found")
		case usecase.ErrUnauthorized:
			h.writeErrorResponse(w, http.StatusForbidden, "access_denied", "Insufficient permissions to update member role")
		default:
			h.writeErrorResponse(w, http.StatusInternalServerError, "update_role_failed", "Failed to update member role")
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Member role successfully updated",
	})
}

// GetUserGroups handles GET /me/groups
func (h *GroupHandler) GetUserGroups(w http.ResponseWriter, r *http.Request) {
	// Get user ID from authentication context
	userID, ok := middleware.GetUserID(r)
	if !ok {
		h.writeErrorResponse(w, http.StatusUnauthorized, "unauthorized", "Authentication required")
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid_user_id", "Invalid user ID")
		return
	}

	// Create get user groups request
	getUserGroupsReq := &usecase.GetUserGroupsRequest{
		UserID: userUUID,
	}

	// Execute get user groups
	result, err := h.groupManagementUseCase.GetUserGroups(r.Context(), getUserGroupsReq)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "groups_fetch_failed", "Failed to fetch user groups")
		return
	}

	// Convert to response format
	groups := make([]GroupResponse, len(result.Groups))
	for i, group := range result.Groups {
		groups[i] = *h.convertToGroupResponse(group, &userUUID, false)
	}

	response := GroupListResponse{
		Groups: groups,
		Total:  len(groups),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// convertToGroupResponse converts domain group to response format
func (h *GroupHandler) convertToGroupResponse(group interface{}, requestingUserID *uuid.UUID, includeMembers bool) *GroupResponse {
	var response *GroupResponse

	// Handle different input types
	switch g := group.(type) {
	case *domain.GroupWithDetails:
		description := ""
		if g.Description != nil {
			description = *g.Description
		}
		response = &GroupResponse{
			ID:          g.ID.String(),
			Name:        g.Name,
			Description: description,
			MemberCount: g.MemberCount,
			CreatedAt:   g.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   g.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Add owner information
		if g.Owner != nil {
			ownerInfo := &UserInfo{
				ID: g.Owner.User.ID.String(),
			}
			if g.Owner.Profile != nil && g.Owner.Profile.DisplayName != nil {
				ownerInfo.DisplayName = g.Owner.Profile.DisplayName
			}
			response.Owner = ownerInfo
		}

		// Add user's role in the group if available
		if g.UserRole != nil {
			response.UserRole = *g.UserRole
		}

		// Add members if requested and available
		if includeMembers && g.Members != nil {
			members := make([]GroupMemberResponse, len(g.Members))
			for i, member := range g.Members {
				members[i] = GroupMemberResponse{
					User: UserInfo{
						ID:          member.UserID.String(),
						DisplayName: nil, // Would need to be fetched from user repository
					},
					Role:     string(member.Role), // Convert GroupRole to string
					JoinedAt: member.JoinedAt.Format("2006-01-02T15:04:05Z07:00"),
				}
			}
			response.Members = members
		}

	case *usecase.CreateGroupResponse:
		description := ""
		if g.Group.Description != nil {
			description = *g.Group.Description
		}
		response = &GroupResponse{
			ID:          g.Group.ID.String(),
			Name:        g.Group.Name,
			Description: description,
			MemberCount: 1, // Owner is the only member initially
			CreatedAt:   g.Group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   g.Group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

		// Set user as owner
		if requestingUserID != nil {
			response.UserRole = string(domain.GroupRoleOwner)
			response.Owner = &UserInfo{
				ID:          requestingUserID.String(),
				DisplayName: nil, // Would need to be fetched from user repository
			}
		}

	case *usecase.UpdateGroupResponse:
		description := ""
		if g.Group.Description != nil {
			description = *g.Group.Description
		}
		response = &GroupResponse{
			ID:          g.Group.ID.String(),
			Name:        g.Group.Name,
			Description: description,
			MemberCount: 0, // Would need to be fetched separately
			CreatedAt:   g.Group.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt:   g.Group.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}

	default:
		// Return empty response for unknown types
		response = &GroupResponse{}
	}

	return response
}

// writeErrorResponse writes a standardized error response
func (h *GroupHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, errorCode, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{
		Error:   errorCode,
		Message: message,
	})
}

// RegisterRoutes registers group management routes with the given router
func (h *GroupHandler) RegisterRoutes(router *mux.Router, authMiddleware *middleware.AuthMiddleware) {
	// All group routes require authentication
	protected := router.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	// Group CRUD operations
	protected.HandleFunc("/groups", h.CreateGroup).Methods("POST")
	protected.HandleFunc("/groups/{id}", h.GetGroup).Methods("GET")
	protected.HandleFunc("/groups/{id}", h.UpdateGroup).Methods("PUT")
	protected.HandleFunc("/groups/{id}", h.DeleteGroup).Methods("DELETE")

	// Group member management
	protected.HandleFunc("/groups/{id}/members", h.AddGroupMember).Methods("POST")
	protected.HandleFunc("/groups/{id}/members/{userId}", h.RemoveGroupMember).Methods("DELETE")
	protected.HandleFunc("/groups/{id}/members/{userId}", h.UpdateMemberRole).Methods("PUT")

	// User's groups
	protected.HandleFunc("/me/groups", h.GetUserGroups).Methods("GET")
}
