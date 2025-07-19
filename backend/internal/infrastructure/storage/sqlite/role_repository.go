package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"

	"github.com/google/uuid"
)

// RoleRepository implements the RoleStorage interface for SQLite
type RoleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sql.DB) interfaces.RoleStorage {
	return &RoleRepository{
		db: db,
	}
}

// CreateRole creates a new role
func (r *RoleRepository) CreateRole(ctx context.Context, role *entities.Role) error {
	if role.ID == "" {
		role.ID = uuid.New().String()
	}

	query := `
		INSERT INTO roles (id, name, display_name, description, is_system, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	role.CreatedAt = now
	role.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		role.ID, role.Name, role.DisplayName, role.Description, role.IsSystem,
		role.CreatedAt, role.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: roles.name") {
			return errors.NewAuthError(errors.CodeRoleExists, "Role already exists", errors.ErrRoleAlreadyExists)
		}
		return fmt.Errorf("failed to create role: %w", err)
	}

	return nil
}

// GetRoleByID retrieves a role by ID
func (r *RoleRepository) GetRoleByID(ctx context.Context, id string) (*entities.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at
		FROM roles
		WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	role, err := r.scanRole(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewAuthError(errors.CodeRoleNotFound, "Role not found", errors.ErrRoleNotFound)
		}
		return nil, fmt.Errorf("failed to get role by ID: %w", err)
	}

	// Load permissions for the role
	permissions, err := r.GetRolePermissions(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}
	role.Permissions = permissions

	return role, nil
}

// GetRoleByName retrieves a role by name
func (r *RoleRepository) GetRoleByName(ctx context.Context, name string) (*entities.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at
		FROM roles
		WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	role, err := r.scanRole(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewAuthError(errors.CodeRoleNotFound, "Role not found", errors.ErrRoleNotFound)
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	// Load permissions for the role
	permissions, err := r.GetRolePermissions(ctx, role.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load role permissions: %w", err)
	}
	role.Permissions = permissions

	return role, nil
}

// GetRoles retrieves roles with filtering
func (r *RoleRepository) GetRoles(ctx context.Context, filter entities.RoleFilter) ([]*entities.Role, error) {
	query := `
		SELECT id, name, display_name, description, is_system, created_at, updated_at
		FROM roles`

	var conditions []string
	var args []interface{}

	if filter.Name != nil {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.IsSystem != nil {
		conditions = append(conditions, "is_system = ?")
		args = append(args, *filter.IsSystem)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY created_at DESC"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		role, err := r.scanRole(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}

		// Load permissions for each role
		permissions, err := r.GetRolePermissions(ctx, role.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to load role permissions: %w", err)
		}
		role.Permissions = permissions

		roles = append(roles, role)
	}

	return roles, nil
}

// UpdateRole updates a role
func (r *RoleRepository) UpdateRole(ctx context.Context, role *entities.Role) error {
	query := `
		UPDATE roles 
		SET name = ?, display_name = ?, description = ?, updated_at = ?
		WHERE id = ?`

	role.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		role.Name, role.DisplayName, role.Description, role.UpdatedAt, role.ID)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: roles.name") {
			return errors.NewAuthError(errors.CodeRoleExists, "Role already exists", errors.ErrRoleAlreadyExists)
		}
		return fmt.Errorf("failed to update role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAuthError(errors.CodeRoleNotFound, "Role not found", errors.ErrRoleNotFound)
	}

	return nil
}

// DeleteRole deletes a role
func (r *RoleRepository) DeleteRole(ctx context.Context, id string) error {
	// Check if role is system role
	role, err := r.GetRoleByID(ctx, id)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return errors.NewAuthError(errors.CodeForbidden, "Cannot delete system role", errors.ErrCannotDeleteSystemRole)
	}

	query := `DELETE FROM roles WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAuthError(errors.CodeRoleNotFound, "Role not found", errors.ErrRoleNotFound)
	}

	return nil
}

// scanRole scans a role row
func (r *RoleRepository) scanRole(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Role, error) {
	var role entities.Role

	err := scanner.Scan(
		&role.ID, &role.Name, &role.DisplayName, &role.Description,
		&role.IsSystem, &role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &role, nil
}

// CreatePermission creates a new permission
func (r *RoleRepository) CreatePermission(ctx context.Context, permission *entities.Permission) error {
	if permission.ID == "" {
		permission.ID = uuid.New().String()
	}

	query := `
		INSERT INTO permissions (id, name, display_name, description, resource, action, is_system, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	permission.CreatedAt = now
	permission.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		permission.ID, permission.Name, permission.DisplayName, permission.Description,
		permission.Resource, permission.Action, permission.IsSystem,
		permission.CreatedAt, permission.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: permissions.name") {
			return errors.NewAuthError(errors.CodePermissionExists, "Permission already exists", errors.ErrPermissionAlreadyExists)
		}
		return fmt.Errorf("failed to create permission: %w", err)
	}

	return nil
}

// GetPermissionByID retrieves a permission by ID
func (r *RoleRepository) GetPermissionByID(ctx context.Context, id string) (*entities.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource, action, is_system, created_at, updated_at
		FROM permissions
		WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	permission, err := r.scanPermission(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewAuthError(errors.CodePermissionNotFound, "Permission not found", errors.ErrPermissionNotFound)
		}
		return nil, fmt.Errorf("failed to get permission by ID: %w", err)
	}

	return permission, nil
}

// GetPermissionByName retrieves a permission by name
func (r *RoleRepository) GetPermissionByName(ctx context.Context, name string) (*entities.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource, action, is_system, created_at, updated_at
		FROM permissions
		WHERE name = ?`

	row := r.db.QueryRowContext(ctx, query, name)
	permission, err := r.scanPermission(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewAuthError(errors.CodePermissionNotFound, "Permission not found", errors.ErrPermissionNotFound)
		}
		return nil, fmt.Errorf("failed to get permission by name: %w", err)
	}

	return permission, nil
}

// GetPermissions retrieves permissions with filtering
func (r *RoleRepository) GetPermissions(ctx context.Context, filter entities.PermissionFilter) ([]*entities.Permission, error) {
	query := `
		SELECT id, name, display_name, description, resource, action, is_system, created_at, updated_at
		FROM permissions`

	var conditions []string
	var args []interface{}

	if filter.Name != nil {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.Resource != nil {
		conditions = append(conditions, "resource = ?")
		args = append(args, *filter.Resource)
	}

	if filter.Action != nil {
		conditions = append(conditions, "action = ?")
		args = append(args, *filter.Action)
	}

	if filter.IsSystem != nil {
		conditions = append(conditions, "is_system = ?")
		args = append(args, *filter.IsSystem)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY resource, action"

	if filter.Limit > 0 {
		query += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	if filter.Offset > 0 {
		query += " OFFSET ?"
		args = append(args, filter.Offset)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*entities.Permission
	for rows.Next() {
		permission, err := r.scanPermission(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// UpdatePermission updates a permission
func (r *RoleRepository) UpdatePermission(ctx context.Context, permission *entities.Permission) error {
	query := `
		UPDATE permissions
		SET name = ?, display_name = ?, description = ?, resource = ?, action = ?, updated_at = ?
		WHERE id = ?`

	permission.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		permission.Name, permission.DisplayName, permission.Description,
		permission.Resource, permission.Action, permission.UpdatedAt, permission.ID)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: permissions.name") {
			return errors.NewAuthError(errors.CodePermissionExists, "Permission already exists", errors.ErrPermissionAlreadyExists)
		}
		return fmt.Errorf("failed to update permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAuthError(errors.CodePermissionNotFound, "Permission not found", errors.ErrPermissionNotFound)
	}

	return nil
}

// DeletePermission deletes a permission
func (r *RoleRepository) DeletePermission(ctx context.Context, id string) error {
	// Check if permission is system permission
	permission, err := r.GetPermissionByID(ctx, id)
	if err != nil {
		return err
	}

	if permission.IsSystem {
		return errors.NewAuthError(errors.CodeForbidden, "Cannot delete system permission", errors.ErrCannotDeleteSystemPermission)
	}

	query := `DELETE FROM permissions WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete permission: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAuthError(errors.CodePermissionNotFound, "Permission not found", errors.ErrPermissionNotFound)
	}

	return nil
}

// scanPermission scans a permission row
func (r *RoleRepository) scanPermission(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.Permission, error) {
	var permission entities.Permission

	err := scanner.Scan(
		&permission.ID, &permission.Name, &permission.DisplayName, &permission.Description,
		&permission.Resource, &permission.Action, &permission.IsSystem,
		&permission.CreatedAt, &permission.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &permission, nil
}

// AssignPermissionToRole assigns a permission to a role
func (r *RoleRepository) AssignPermissionToRole(ctx context.Context, roleID, permissionID string) error {
	query := `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES (?, ?, ?)`

	_, err := r.db.ExecContext(ctx, query, roleID, permissionID, time.Now())
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			// Permission already assigned to role, ignore
			return nil
		}
		return fmt.Errorf("failed to assign permission to role: %w", err)
	}

	return nil
}

// RemovePermissionFromRole removes a permission from a role
func (r *RoleRepository) RemovePermissionFromRole(ctx context.Context, roleID, permissionID string) error {
	query := `DELETE FROM role_permissions WHERE role_id = ? AND permission_id = ?`

	result, err := r.db.ExecContext(ctx, query, roleID, permissionID)
	if err != nil {
		return fmt.Errorf("failed to remove permission from role: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAuthError(errors.CodePermissionNotFound, "Permission not assigned to role", errors.ErrPermissionNotFound)
	}

	return nil
}

// GetRolePermissions retrieves all permissions for a role
func (r *RoleRepository) GetRolePermissions(ctx context.Context, roleID string) ([]*entities.Permission, error) {
	query := `
		SELECT p.id, p.name, p.display_name, p.description, p.resource, p.action, p.is_system, p.created_at, p.updated_at
		FROM permissions p
		INNER JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = ?
		ORDER BY p.resource, p.action`

	rows, err := r.db.QueryContext(ctx, query, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to query role permissions: %w", err)
	}
	defer rows.Close()

	var permissions []*entities.Permission
	for rows.Next() {
		permission, err := r.scanPermission(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan permission: %w", err)
		}
		permissions = append(permissions, permission)
	}

	return permissions, nil
}

// GetPermissionRoles retrieves all roles that have a specific permission
func (r *RoleRepository) GetPermissionRoles(ctx context.Context, permissionID string) ([]*entities.Role, error) {
	query := `
		SELECT r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM roles r
		INNER JOIN role_permissions rp ON r.id = rp.role_id
		WHERE rp.permission_id = ?
		ORDER BY r.name`

	rows, err := r.db.QueryContext(ctx, query, permissionID)
	if err != nil {
		return nil, fmt.Errorf("failed to query permission roles: %w", err)
	}
	defer rows.Close()

	var roles []*entities.Role
	for rows.Next() {
		role, err := r.scanRole(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	return roles, nil
}

// SetRolePermissions sets all permissions for a role (replaces existing permissions)
func (r *RoleRepository) SetRolePermissions(ctx context.Context, roleID string, permissionIDs []string) error {
	// Start transaction
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Remove all existing permissions for the role
	_, err = tx.ExecContext(ctx, "DELETE FROM role_permissions WHERE role_id = ?", roleID)
	if err != nil {
		return fmt.Errorf("failed to remove existing permissions: %w", err)
	}

	// Add new permissions
	if len(permissionIDs) > 0 {
		query := "INSERT INTO role_permissions (role_id, permission_id, created_at) VALUES "
		values := make([]string, len(permissionIDs))
		args := make([]interface{}, len(permissionIDs)*3)

		now := time.Now()
		for i, permissionID := range permissionIDs {
			values[i] = "(?, ?, ?)"
			args[i*3] = roleID
			args[i*3+1] = permissionID
			args[i*3+2] = now
		}

		query += strings.Join(values, ", ")
		_, err = tx.ExecContext(ctx, query, args...)
		if err != nil {
			return fmt.Errorf("failed to add new permissions: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// HasPermission checks if a role has a specific permission
func (r *RoleRepository) HasPermission(ctx context.Context, roleID, permissionName string) (bool, error) {
	query := `
		SELECT 1
		FROM role_permissions rp
		INNER JOIN permissions p ON rp.permission_id = p.id
		WHERE rp.role_id = ? AND (p.name = ? OR p.name = '*')
		LIMIT 1`

	row := r.db.QueryRowContext(ctx, query, roleID, permissionName)
	var exists int
	err := row.Scan(&exists)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, fmt.Errorf("failed to check permission: %w", err)
	}

	return true, nil
}

// Health checks the health of the role repository
func (r *RoleRepository) Health(ctx context.Context) error {
	query := `SELECT 1 FROM roles LIMIT 1`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
