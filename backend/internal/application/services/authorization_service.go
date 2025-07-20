package services

import (
	"context"
	"strings"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"
)

// AuthorizationService implements authorization operations
type AuthorizationService struct {
	userStorage interfaces.UserStorage
	roleStorage interfaces.RoleStorage
}

// NewAuthorizationService creates a new authorization service
func NewAuthorizationService(userStorage interfaces.UserStorage, roleStorage interfaces.RoleStorage) interfaces.AuthorizationService {
	return &AuthorizationService{
		userStorage: userStorage,
		roleStorage: roleStorage,
	}
}

// HasPermission checks if a user has a specific permission
func (s *AuthorizationService) HasPermission(ctx context.Context, userID, permission string) (bool, error) {
	if userID == "" || permission == "" {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "user ID and permission cannot be empty", nil)
	}

	// Get user with roles
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return false, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Check if user is active
	if user.Status != entities.UserStatusActive {
		return false, errors.NewAuthError(errors.CodeUserInactive, "user is not active", nil)
	}

	// Get user's role permissions
	permissions, err := s.roleStorage.GetRolePermissions(ctx, user.RoleID)
	if err != nil {
		return false, errors.NewAuthError("INTERNAL_ERROR", "failed to get role permissions", err)
	}

	// Check if user has the required permission
	for _, perm := range permissions {
		if s.matchesPermission(perm.Name, permission) {
			return true, nil
		}
	}

	return false, nil
}

// HasAnyPermission checks if a user has any of the specified permissions
func (s *AuthorizationService) HasAnyPermission(ctx context.Context, userID string, permissions []string) (bool, error) {
	if userID == "" || len(permissions) == 0 {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "user ID and permissions cannot be empty", nil)
	}

	for _, permission := range permissions {
		hasPermission, err := s.HasPermission(ctx, userID, permission)
		if err != nil {
			continue // Continue checking other permissions
		}
		if hasPermission {
			return true, nil
		}
	}

	return false, nil
}

// HasAllPermissions checks if a user has all of the specified permissions
func (s *AuthorizationService) HasAllPermissions(ctx context.Context, userID string, permissions []string) (bool, error) {
	if userID == "" || len(permissions) == 0 {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "user ID and permissions cannot be empty", nil)
	}

	for _, permission := range permissions {
		hasPermission, err := s.HasPermission(ctx, userID, permission)
		if err != nil {
			return false, err
		}
		if !hasPermission {
			return false, nil
		}
	}

	return true, nil
}

// HasRole checks if a user has a specific role
func (s *AuthorizationService) HasRole(ctx context.Context, userID, roleName string) (bool, error) {
	if userID == "" || roleName == "" {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "user ID and role name cannot be empty", nil)
	}

	// Get user with roles
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return false, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Check if user is active
	if user.Status != entities.UserStatusActive {
		return false, errors.NewAuthError(errors.CodeUserInactive, "user is not active", nil)
	}

	// Get user's role
	role, err := s.roleStorage.GetRoleByID(ctx, user.RoleID)
	if err != nil {
		return false, errors.NewAuthError(errors.CodeRoleNotFound, "role not found", err)
	}

	// Check if user has the role
	return role.Name == roleName, nil
}

// HasAnyRole checks if a user has any of the specified roles
func (s *AuthorizationService) HasAnyRole(ctx context.Context, userID string, roleNames []string) (bool, error) {
	if userID == "" || len(roleNames) == 0 {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "user ID and role names cannot be empty", nil)
	}

	for _, roleName := range roleNames {
		hasRole, err := s.HasRole(ctx, userID, roleName)
		if err != nil {
			continue // Continue checking other roles
		}
		if hasRole {
			return true, nil
		}
	}

	return false, nil
}

// CanAccessResource checks if a user can access a resource with a specific action
func (s *AuthorizationService) CanAccessResource(ctx context.Context, userID, resource, action string) (bool, error) {
	if userID == "" || resource == "" || action == "" {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "user ID, resource, and action cannot be empty", nil)
	}

	// Build permission string in format "resource:action"
	permission := resource + ":" + action

	// Check if user has the specific permission
	hasPermission, err := s.HasPermission(ctx, userID, permission)
	if err != nil {
		return false, err
	}
	if hasPermission {
		return true, nil
	}

	// Check for wildcard permissions
	wildcardPermissions := []string{
		resource + ":*", // resource:*
		"*:" + action,   // *:action
		"*:*",           // *:*
	}

	return s.HasAnyPermission(ctx, userID, wildcardPermissions)
}

// CanModifyUser checks if an actor can modify a target user
func (s *AuthorizationService) CanModifyUser(ctx context.Context, actorID, targetUserID string) (bool, error) {
	if actorID == "" || targetUserID == "" {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "actor ID and target user ID cannot be empty", nil)
	}

	// Users can always modify themselves (with some restrictions)
	if actorID == targetUserID {
		return true, nil
	}

	// Check if actor has user management permissions
	canManageUsers, err := s.HasPermission(ctx, actorID, "users:write")
	if err != nil {
		return false, err
	}
	if canManageUsers {
		return true, nil
	}

	// Check if actor is an admin
	isAdmin, err := s.HasRole(ctx, actorID, "admin")
	if err != nil {
		return false, err
	}
	if isAdmin {
		return true, nil
	}

	return false, nil
}

// CanManageRole checks if a user can manage a specific role
func (s *AuthorizationService) CanManageRole(ctx context.Context, userID, roleID string) (bool, error) {
	if userID == "" || roleID == "" {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "user ID and role ID cannot be empty", nil)
	}

	// Get the role to check if it's a system role
	role, err := s.roleStorage.GetRoleByID(ctx, roleID)
	if err != nil {
		return false, errors.NewAuthError(errors.CodeRoleNotFound, "role not found", err)
	}

	// System roles cannot be managed by regular users
	if role.IsSystem {
		// Only super admins can manage system roles
		isSuperAdmin, err := s.HasRole(ctx, userID, "super_admin")
		if err != nil {
			return false, err
		}
		return isSuperAdmin, nil
	}

	// Check if user has role management permissions
	canManageRoles, err := s.HasPermission(ctx, userID, "roles:write")
	if err != nil {
		return false, err
	}
	if canManageRoles {
		return true, nil
	}

	// Check if user is an admin
	isAdmin, err := s.HasRole(ctx, userID, "admin")
	if err != nil {
		return false, err
	}

	return isAdmin, nil
}

// CanAPIKeyAccess checks if an API key can access a resource with a specific action
func (s *AuthorizationService) CanAPIKeyAccess(ctx context.Context, apiKey *entities.APIKey, resource, action string) (bool, error) {
	if apiKey == nil || resource == "" || action == "" {
		return false, errors.NewAuthError(errors.CodeInvalidInput, "API key, resource, and action cannot be empty", nil)
	}

	// Check if API key is active
	if !apiKey.IsActive {
		return false, errors.NewAuthError("API_KEY_REVOKED", "API key is revoked", nil)
	}

	// Build permission string in format "resource:action"
	permission := resource + ":" + action

	// Check if API key has the specific permission
	for _, keyPermission := range apiKey.Permissions {
		if keyPermission == permission {
			return true, nil
		}

		// Check for wildcard permissions
		if s.matchesWildcardPermission(keyPermission, resource, action) {
			return true, nil
		}
	}

	return false, nil
}

// matchesWildcardPermission checks if a wildcard permission matches the resource and action
func (s *AuthorizationService) matchesWildcardPermission(permission, resource, action string) bool {
	parts := strings.Split(permission, ":")
	if len(parts) != 2 {
		return false
	}

	permResource, permAction := parts[0], parts[1]

	// Check resource match
	resourceMatch := permResource == "*" || permResource == resource

	// Check action match
	actionMatch := permAction == "*" || permAction == action

	return resourceMatch && actionMatch
}

// GetUserPermissions returns all permissions for a user (helper method)
func (s *AuthorizationService) GetUserPermissions(ctx context.Context, userID string) ([]string, error) {
	if userID == "" {
		return nil, errors.NewAuthError(errors.CodeInvalidInput, "user ID cannot be empty", nil)
	}

	// Get user with roles
	user, err := s.userStorage.GetUserByID(ctx, userID)
	if err != nil {
		return nil, errors.NewAuthError(errors.CodeUserNotFound, "user not found", err)
	}

	// Check if user is active
	if user.Status != entities.UserStatusActive {
		return nil, errors.NewAuthError(errors.CodeUserInactive, "user is not active", nil)
	}

	// Get permissions from user's role
	var permissions []string
	rolePermissions, err := s.roleStorage.GetRolePermissions(ctx, user.RoleID)
	if err != nil {
		return nil, errors.NewAuthError("INTERNAL_ERROR", "failed to get role permissions", err)
	}

	for _, permission := range rolePermissions {
		permissionKey := permission.Resource + ":" + permission.Action
		permissions = append(permissions, permissionKey)
	}

	return permissions, nil
}

// matchesPermission checks if a permission matches the required permission with wildcard support
func (s *AuthorizationService) matchesPermission(permissionName, requiredPermission string) bool {
	// Direct match
	if permissionName == requiredPermission {
		return true
	}

	// Wildcard match - "*:*" grants all permissions
	if permissionName == "*:*" {
		return true
	}

	// Parse permission strings
	permParts := strings.Split(permissionName, ":")
	reqParts := strings.Split(requiredPermission, ":")

	if len(permParts) != 2 || len(reqParts) != 2 {
		return false
	}

	// Resource wildcard - "resource:*" grants all actions on resource
	if permParts[0] == reqParts[0] && permParts[1] == "*" {
		return true
	}

	// Action wildcard - "*:action" grants action on all resources
	if permParts[0] == "*" && permParts[1] == reqParts[1] {
		return true
	}

	return false
}
