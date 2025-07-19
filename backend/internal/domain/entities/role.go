package entities

import (
	"time"
)

// Role represents a role in the system
type Role struct {
	ID          string        `json:"id" db:"id"`
	Name        string        `json:"name" db:"name"`
	DisplayName string        `json:"display_name" db:"display_name"`
	Description string        `json:"description" db:"description"`
	Permissions []*Permission `json:"permissions,omitempty" db:"-"` // Loaded separately
	IsSystem    bool          `json:"is_system" db:"is_system"`     // System roles cannot be deleted
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" db:"updated_at"`
}

// HasPermission checks if the role has a specific permission
func (r *Role) HasPermission(permission string) bool {
	for _, p := range r.Permissions {
		if p.Name == permission || p.Name == "*" {
			return true
		}
	}
	return false
}

// GetPermissionNames returns a slice of permission names
func (r *Role) GetPermissionNames() []string {
	names := make([]string, len(r.Permissions))
	for i, p := range r.Permissions {
		names[i] = p.Name
	}
	return names
}

// Permission represents a permission in the system
type Permission struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	DisplayName string    `json:"display_name" db:"display_name"`
	Description string    `json:"description" db:"description"`
	Resource    string    `json:"resource" db:"resource"` // e.g., "instruments", "data", "users"
	Action      string    `json:"action" db:"action"`     // e.g., "read", "write", "delete"
	IsSystem    bool      `json:"is_system" db:"is_system"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// RolePermission represents the many-to-many relationship between roles and permissions
type RolePermission struct {
	RoleID       string    `json:"role_id" db:"role_id"`
	PermissionID string    `json:"permission_id" db:"permission_id"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
}

// Predefined system roles
const (
	RoleAdmin     = "admin"
	RoleUser      = "user"
	RoleReadOnly  = "readonly"
	RoleAPIAccess = "api_access"
)

// Predefined permissions
const (
	// User management permissions
	PermissionUsersRead   = "users:read"
	PermissionUsersWrite  = "users:write"
	PermissionUsersDelete = "users:delete"

	// Role management permissions
	PermissionRolesRead   = "roles:read"
	PermissionRolesWrite  = "roles:write"
	PermissionRolesDelete = "roles:delete"

	// Instrument management permissions
	PermissionInstrumentsRead   = "instruments:read"
	PermissionInstrumentsWrite  = "instruments:write"
	PermissionInstrumentsDelete = "instruments:delete"

	// Data access permissions
	PermissionDataRead   = "data:read"
	PermissionDataWrite  = "data:write"
	PermissionDataDelete = "data:delete"

	// Subscription management permissions
	PermissionSubscriptionsRead   = "subscriptions:read"
	PermissionSubscriptionsWrite  = "subscriptions:write"
	PermissionSubscriptionsDelete = "subscriptions:delete"

	// API key management permissions
	PermissionAPIKeysRead   = "api_keys:read"
	PermissionAPIKeysWrite  = "api_keys:write"
	PermissionAPIKeysDelete = "api_keys:delete"

	// System administration permissions
	PermissionSystemRead  = "system:read"
	PermissionSystemWrite = "system:write"

	// Special permissions
	PermissionAll = "*" // Super admin permission
)

// CreateRoleRequest represents a request to create a new role
type CreateRoleRequest struct {
	Name        string   `json:"name" validate:"required,min=1,max=50"`
	DisplayName string   `json:"display_name" validate:"required,min=1,max=100"`
	Description string   `json:"description" validate:"max=500"`
	Permissions []string `json:"permissions" validate:"required,min=1"`
}

// UpdateRoleRequest represents a request to update a role
type UpdateRoleRequest struct {
	DisplayName *string  `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Permissions []string `json:"permissions,omitempty"`
}

// CreatePermissionRequest represents a request to create a new permission
type CreatePermissionRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=100"`
	Description string `json:"description" validate:"max=500"`
	Resource    string `json:"resource" validate:"required,min=1,max=50"`
	Action      string `json:"action" validate:"required,min=1,max=50"`
}

// UpdatePermissionRequest represents a request to update a permission
type UpdatePermissionRequest struct {
	DisplayName *string `json:"display_name,omitempty" validate:"omitempty,min=1,max=100"`
	Description *string `json:"description,omitempty" validate:"omitempty,max=500"`
	Resource    *string `json:"resource,omitempty" validate:"omitempty,min=1,max=50"`
	Action      *string `json:"action,omitempty" validate:"omitempty,min=1,max=50"`
}

// RoleFilter represents filters for querying roles
type RoleFilter struct {
	Name     *string `json:"name,omitempty"`
	IsSystem *bool   `json:"is_system,omitempty"`
	Limit    int     `json:"limit,omitempty"`
	Offset   int     `json:"offset,omitempty"`
}

// PermissionFilter represents filters for querying permissions
type PermissionFilter struct {
	Name     *string `json:"name,omitempty"`
	Resource *string `json:"resource,omitempty"`
	Action   *string `json:"action,omitempty"`
	IsSystem *bool   `json:"is_system,omitempty"`
	Limit    int     `json:"limit,omitempty"`
	Offset   int     `json:"offset,omitempty"`
}

// GetDefaultRoles returns the default system roles
func GetDefaultRoles() []Role {
	now := time.Now()
	return []Role{
		{
			ID:          "admin",
			Name:        RoleAdmin,
			DisplayName: "Administrator",
			Description: "Full system access with all permissions",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "user",
			Name:        RoleUser,
			DisplayName: "User",
			Description: "Standard user with basic data access",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "readonly",
			Name:        RoleReadOnly,
			DisplayName: "Read Only",
			Description: "Read-only access to data and instruments",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "api_access",
			Name:        RoleAPIAccess,
			DisplayName: "API Access",
			Description: "Programmatic API access for external systems",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
}

// GetDefaultPermissions returns the default system permissions
func GetDefaultPermissions() []Permission {
	now := time.Now()
	return []Permission{
		// User management
		{ID: "users_read", Name: PermissionUsersRead, DisplayName: "Read Users", Description: "View user information", Resource: "users", Action: "read", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "users_write", Name: PermissionUsersWrite, DisplayName: "Write Users", Description: "Create and update users", Resource: "users", Action: "write", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "users_delete", Name: PermissionUsersDelete, DisplayName: "Delete Users", Description: "Delete users", Resource: "users", Action: "delete", IsSystem: true, CreatedAt: now, UpdatedAt: now},

		// Role management
		{ID: "roles_read", Name: PermissionRolesRead, DisplayName: "Read Roles", Description: "View role information", Resource: "roles", Action: "read", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "roles_write", Name: PermissionRolesWrite, DisplayName: "Write Roles", Description: "Create and update roles", Resource: "roles", Action: "write", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "roles_delete", Name: PermissionRolesDelete, DisplayName: "Delete Roles", Description: "Delete roles", Resource: "roles", Action: "delete", IsSystem: true, CreatedAt: now, UpdatedAt: now},

		// Instrument management
		{ID: "instruments_read", Name: PermissionInstrumentsRead, DisplayName: "Read Instruments", Description: "View instrument information", Resource: "instruments", Action: "read", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "instruments_write", Name: PermissionInstrumentsWrite, DisplayName: "Write Instruments", Description: "Create and update instruments", Resource: "instruments", Action: "write", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "instruments_delete", Name: PermissionInstrumentsDelete, DisplayName: "Delete Instruments", Description: "Delete instruments", Resource: "instruments", Action: "delete", IsSystem: true, CreatedAt: now, UpdatedAt: now},

		// Data access
		{ID: "data_read", Name: PermissionDataRead, DisplayName: "Read Data", Description: "Access market data", Resource: "data", Action: "read", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "data_write", Name: PermissionDataWrite, DisplayName: "Write Data", Description: "Store market data", Resource: "data", Action: "write", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "data_delete", Name: PermissionDataDelete, DisplayName: "Delete Data", Description: "Delete market data", Resource: "data", Action: "delete", IsSystem: true, CreatedAt: now, UpdatedAt: now},

		// Subscription management
		{ID: "subscriptions_read", Name: PermissionSubscriptionsRead, DisplayName: "Read Subscriptions", Description: "View subscriptions", Resource: "subscriptions", Action: "read", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "subscriptions_write", Name: PermissionSubscriptionsWrite, DisplayName: "Write Subscriptions", Description: "Manage subscriptions", Resource: "subscriptions", Action: "write", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "subscriptions_delete", Name: PermissionSubscriptionsDelete, DisplayName: "Delete Subscriptions", Description: "Delete subscriptions", Resource: "subscriptions", Action: "delete", IsSystem: true, CreatedAt: now, UpdatedAt: now},

		// API key management
		{ID: "api_keys_read", Name: PermissionAPIKeysRead, DisplayName: "Read API Keys", Description: "View API keys", Resource: "api_keys", Action: "read", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "api_keys_write", Name: PermissionAPIKeysWrite, DisplayName: "Write API Keys", Description: "Create and update API keys", Resource: "api_keys", Action: "write", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "api_keys_delete", Name: PermissionAPIKeysDelete, DisplayName: "Delete API Keys", Description: "Delete API keys", Resource: "api_keys", Action: "delete", IsSystem: true, CreatedAt: now, UpdatedAt: now},

		// System administration
		{ID: "system_read", Name: PermissionSystemRead, DisplayName: "Read System", Description: "View system information", Resource: "system", Action: "read", IsSystem: true, CreatedAt: now, UpdatedAt: now},
		{ID: "system_write", Name: PermissionSystemWrite, DisplayName: "Write System", Description: "Modify system settings", Resource: "system", Action: "write", IsSystem: true, CreatedAt: now, UpdatedAt: now},

		// Super admin
		{ID: "all", Name: PermissionAll, DisplayName: "All Permissions", Description: "Full system access", Resource: "*", Action: "*", IsSystem: true, CreatedAt: now, UpdatedAt: now},
	}
}

// GetDefaultRolePermissions returns the default role-permission mappings
func GetDefaultRolePermissions() map[string][]string {
	return map[string][]string{
		RoleAdmin: {PermissionAll},
		RoleUser: {
			PermissionInstrumentsRead,
			PermissionDataRead,
			PermissionSubscriptionsRead,
			PermissionSubscriptionsWrite,
			PermissionAPIKeysRead,
			PermissionAPIKeysWrite,
		},
		RoleReadOnly: {
			PermissionInstrumentsRead,
			PermissionDataRead,
			PermissionSubscriptionsRead,
		},
		RoleAPIAccess: {
			PermissionInstrumentsRead,
			PermissionDataRead,
			PermissionDataWrite,
			PermissionSubscriptionsRead,
			PermissionSubscriptionsWrite,
		},
	}
}
