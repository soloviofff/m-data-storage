package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"m-data-storage/internal/domain/entities"
	"m-data-storage/internal/domain/errors"
	"m-data-storage/internal/domain/interfaces"

	"github.com/google/uuid"
)

// UserRepository implements the UserStorage interface for SQLite
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) interfaces.UserStorage {
	return &UserRepository{
		db: db,
	}
}

// CreateUser creates a new user
func (r *UserRepository) CreateUser(ctx context.Context, user *entities.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}

	query := `
		INSERT INTO users (id, username, email, password_hash, first_name, last_name, status, role_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.db.ExecContext(ctx, query,
		user.ID, user.Username, user.Email, user.PasswordHash,
		user.FirstName, user.LastName, user.Status, user.RoleID,
		user.CreatedAt, user.UpdatedAt)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			return errors.NewUserExistsError(user.Username, "")
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return errors.NewUserExistsError("", user.Email)
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetUserByID retrieves a user by ID
func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*entities.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.first_name, u.last_name, 
		       u.status, u.role_id, u.created_at, u.updated_at, u.last_login_at,
		       r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	user, err := r.scanUserWithRole(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewUserNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

// GetUserByUsername retrieves a user by username
func (r *UserRepository) GetUserByUsername(ctx context.Context, username string) (*entities.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.first_name, u.last_name, 
		       u.status, u.role_id, u.created_at, u.updated_at, u.last_login_at,
		       r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username = ?`

	row := r.db.QueryRowContext(ctx, query, username)
	user, err := r.scanUserWithRole(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewUserNotFoundError(username)
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}

	return user, nil
}

// GetUserByEmail retrieves a user by email
func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entities.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.first_name, u.last_name, 
		       u.status, u.role_id, u.created_at, u.updated_at, u.last_login_at,
		       r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.email = ?`

	row := r.db.QueryRowContext(ctx, query, email)
	user, err := r.scanUserWithRole(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewUserNotFoundError(email)
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return user, nil
}

// GetUsers retrieves users with filtering
func (r *UserRepository) GetUsers(ctx context.Context, filter entities.UserFilter) ([]*entities.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.first_name, u.last_name, 
		       u.status, u.role_id, u.created_at, u.updated_at, u.last_login_at,
		       r.id, r.name, r.display_name, r.description, r.is_system, r.created_at, r.updated_at
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id`

	var conditions []string
	var args []interface{}

	if filter.Username != nil {
		conditions = append(conditions, "u.username LIKE ?")
		args = append(args, "%"+*filter.Username+"%")
	}

	if filter.Email != nil {
		conditions = append(conditions, "u.email LIKE ?")
		args = append(args, "%"+*filter.Email+"%")
	}

	if filter.Status != nil {
		conditions = append(conditions, "u.status = ?")
		args = append(args, *filter.Status)
	}

	if filter.RoleID != nil {
		conditions = append(conditions, "u.role_id = ?")
		args = append(args, *filter.RoleID)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY u.created_at DESC"

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
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []*entities.User
	for rows.Next() {
		user, err := r.scanUserWithRole(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

// UpdateUser updates a user
func (r *UserRepository) UpdateUser(ctx context.Context, user *entities.User) error {
	query := `
		UPDATE users 
		SET username = ?, email = ?, password_hash = ?, first_name = ?, last_name = ?, 
		    status = ?, role_id = ?, updated_at = ?
		WHERE id = ?`

	user.UpdatedAt = time.Now()

	result, err := r.db.ExecContext(ctx, query,
		user.Username, user.Email, user.PasswordHash, user.FirstName, user.LastName,
		user.Status, user.RoleID, user.UpdatedAt, user.ID)

	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.username") {
			return errors.NewUserExistsError(user.Username, "")
		}
		if strings.Contains(err.Error(), "UNIQUE constraint failed: users.email") {
			return errors.NewUserExistsError("", user.Email)
		}
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundError(user.ID)
	}

	return nil
}

// DeleteUser deletes a user
func (r *UserRepository) DeleteUser(ctx context.Context, id string) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundError(id)
	}

	return nil
}

// UpdateLastLogin updates the last login time for a user
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID string) error {
	query := `UPDATE users SET last_login_at = ? WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewUserNotFoundError(userID)
	}

	return nil
}

// scanUserWithRole scans a user row with role information
func (r *UserRepository) scanUserWithRole(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.User, error) {
	var user entities.User
	var role entities.Role
	var roleID, roleName, roleDisplayName, roleDescription sql.NullString
	var roleIsSystem sql.NullBool
	var roleCreatedAt, roleUpdatedAt sql.NullTime
	var lastLoginAt sql.NullTime

	err := scanner.Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash,
		&user.FirstName, &user.LastName, &user.Status, &user.RoleID,
		&user.CreatedAt, &user.UpdatedAt, &lastLoginAt,
		&roleID, &roleName, &roleDisplayName, &roleDescription,
		&roleIsSystem, &roleCreatedAt, &roleUpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	if lastLoginAt.Valid {
		user.LastLoginAt = &lastLoginAt.Time
	}

	// Set role if it exists
	if roleID.Valid {
		role.ID = roleID.String
		role.Name = roleName.String
		role.DisplayName = roleDisplayName.String
		role.Description = roleDescription.String
		role.IsSystem = roleIsSystem.Bool
		role.CreatedAt = roleCreatedAt.Time
		role.UpdatedAt = roleUpdatedAt.Time
		user.Role = &role
	}

	return &user, nil
}

// CreateSession creates a new user session
func (r *UserRepository) CreateSession(ctx context.Context, session *entities.UserSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}

	query := `
		INSERT INTO user_sessions (id, user_id, token_hash, refresh_token, expires_at, created_at, last_used_at, ip_address, user_agent, is_revoked)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	session.CreatedAt = now
	session.LastUsedAt = now

	_, err := r.db.ExecContext(ctx, query,
		session.ID, session.UserID, session.TokenHash, session.RefreshToken,
		session.ExpiresAt, session.CreatedAt, session.LastUsedAt,
		session.IPAddress, session.UserAgent, session.IsRevoked)

	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}

	return nil
}

// GetSessionByID retrieves a session by ID
func (r *UserRepository) GetSessionByID(ctx context.Context, id string) (*entities.UserSession, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token, expires_at, created_at, last_used_at, ip_address, user_agent, is_revoked
		FROM user_sessions
		WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	session, err := r.scanSession(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewAuthError(errors.CodeSessionNotFound, "Session not found", errors.ErrSessionNotFound)
		}
		return nil, fmt.Errorf("failed to get session by ID: %w", err)
	}

	return session, nil
}

// GetSessionsByUserID retrieves all sessions for a user
func (r *UserRepository) GetSessionsByUserID(ctx context.Context, userID string) ([]*entities.UserSession, error) {
	query := `
		SELECT id, user_id, token_hash, refresh_token, expires_at, created_at, last_used_at, ip_address, user_agent, is_revoked
		FROM user_sessions
		WHERE user_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*entities.UserSession
	for rows.Next() {
		session, err := r.scanSession(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	return sessions, nil
}

// UpdateSession updates a session
func (r *UserRepository) UpdateSession(ctx context.Context, session *entities.UserSession) error {
	query := `
		UPDATE user_sessions
		SET last_used_at = ?, ip_address = ?, user_agent = ?, is_revoked = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		session.LastUsedAt, session.IPAddress, session.UserAgent, session.IsRevoked, session.ID)

	if err != nil {
		return fmt.Errorf("failed to update session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAuthError(errors.CodeSessionNotFound, "Session not found", errors.ErrSessionNotFound)
	}

	return nil
}

// RevokeSession revokes a session
func (r *UserRepository) RevokeSession(ctx context.Context, id string) error {
	query := `UPDATE user_sessions SET is_revoked = TRUE WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAuthError(errors.CodeSessionNotFound, "Session not found", errors.ErrSessionNotFound)
	}

	return nil
}

// RevokeUserSessions revokes all sessions for a user
func (r *UserRepository) RevokeUserSessions(ctx context.Context, userID string) error {
	query := `UPDATE user_sessions SET is_revoked = TRUE WHERE user_id = ?`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke user sessions: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions
func (r *UserRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `DELETE FROM user_sessions WHERE expires_at < ? OR is_revoked = TRUE`

	_, err := r.db.ExecContext(ctx, query, time.Now())
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	return nil
}

// scanSession scans a session row
func (r *UserRepository) scanSession(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.UserSession, error) {
	var session entities.UserSession

	err := scanner.Scan(
		&session.ID, &session.UserID, &session.TokenHash, &session.RefreshToken,
		&session.ExpiresAt, &session.CreatedAt, &session.LastUsedAt,
		&session.IPAddress, &session.UserAgent, &session.IsRevoked,
	)

	if err != nil {
		return nil, err
	}

	return &session, nil
}

// CreateAPIKey creates a new API key
func (r *UserRepository) CreateAPIKey(ctx context.Context, apiKey *entities.APIKey) error {
	if apiKey.ID == "" {
		apiKey.ID = uuid.New().String()
	}

	// Convert permissions slice to JSON
	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		INSERT INTO api_keys (id, user_id, name, key_hash, prefix, permissions, expires_at, created_at, last_used_at, is_active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	now := time.Now()
	apiKey.CreatedAt = now

	_, err = r.db.ExecContext(ctx, query,
		apiKey.ID, apiKey.UserID, apiKey.Name, apiKey.KeyHash, apiKey.Prefix,
		string(permissionsJSON), apiKey.ExpiresAt, apiKey.CreatedAt, apiKey.LastUsedAt, apiKey.IsActive)

	if err != nil {
		return fmt.Errorf("failed to create API key: %w", err)
	}

	return nil
}

// GetAPIKeyByID retrieves an API key by ID
func (r *UserRepository) GetAPIKeyByID(ctx context.Context, id string) (*entities.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, prefix, permissions, expires_at, created_at, last_used_at, is_active
		FROM api_keys
		WHERE id = ?`

	row := r.db.QueryRowContext(ctx, query, id)
	apiKey, err := r.scanAPIKey(row)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewAPIKeyNotFoundError(id)
		}
		return nil, fmt.Errorf("failed to get API key by ID: %w", err)
	}

	return apiKey, nil
}

// GetAPIKeysByUserID retrieves all API keys for a user
func (r *UserRepository) GetAPIKeysByUserID(ctx context.Context, userID string) ([]*entities.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, prefix, permissions, expires_at, created_at, last_used_at, is_active
		FROM api_keys
		WHERE user_id = ?
		ORDER BY created_at DESC`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*entities.APIKey
	for rows.Next() {
		apiKey, err := r.scanAPIKey(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// GetAPIKeys retrieves API keys with filtering
func (r *UserRepository) GetAPIKeys(ctx context.Context, filter entities.APIKeyFilter) ([]*entities.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_hash, prefix, permissions, expires_at, created_at, last_used_at, is_active
		FROM api_keys`

	var conditions []string
	var args []interface{}

	if filter.UserID != nil {
		conditions = append(conditions, "user_id = ?")
		args = append(args, *filter.UserID)
	}

	if filter.Name != nil {
		conditions = append(conditions, "name LIKE ?")
		args = append(args, "%"+*filter.Name+"%")
	}

	if filter.IsActive != nil {
		conditions = append(conditions, "is_active = ?")
		args = append(args, *filter.IsActive)
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
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var apiKeys []*entities.APIKey
	for rows.Next() {
		apiKey, err := r.scanAPIKey(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		apiKeys = append(apiKeys, apiKey)
	}

	return apiKeys, nil
}

// UpdateAPIKey updates an API key
func (r *UserRepository) UpdateAPIKey(ctx context.Context, apiKey *entities.APIKey) error {
	// Convert permissions slice to JSON
	permissionsJSON, err := json.Marshal(apiKey.Permissions)
	if err != nil {
		return fmt.Errorf("failed to marshal permissions: %w", err)
	}

	query := `
		UPDATE api_keys
		SET name = ?, permissions = ?, expires_at = ?, last_used_at = ?, is_active = ?
		WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query,
		apiKey.Name, string(permissionsJSON), apiKey.ExpiresAt, apiKey.LastUsedAt, apiKey.IsActive, apiKey.ID)

	if err != nil {
		return fmt.Errorf("failed to update API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAPIKeyNotFoundError(apiKey.ID)
	}

	return nil
}

// DeleteAPIKey deletes an API key
func (r *UserRepository) DeleteAPIKey(ctx context.Context, id string) error {
	query := `DELETE FROM api_keys WHERE id = ?`

	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAPIKeyNotFoundError(id)
	}

	return nil
}

// UpdateAPIKeyLastUsed updates the last used time for an API key
func (r *UserRepository) UpdateAPIKeyLastUsed(ctx context.Context, id string) error {
	query := `UPDATE api_keys SET last_used_at = ? WHERE id = ?`

	now := time.Now()
	result, err := r.db.ExecContext(ctx, query, now, id)
	if err != nil {
		return fmt.Errorf("failed to update API key last used: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return errors.NewAPIKeyNotFoundError(id)
	}

	return nil
}

// scanAPIKey scans an API key row
func (r *UserRepository) scanAPIKey(scanner interface {
	Scan(dest ...interface{}) error
}) (*entities.APIKey, error) {
	var apiKey entities.APIKey
	var permissionsJSON string
	var expiresAt, lastUsedAt sql.NullTime

	err := scanner.Scan(
		&apiKey.ID, &apiKey.UserID, &apiKey.Name, &apiKey.KeyHash, &apiKey.Prefix,
		&permissionsJSON, &expiresAt, &apiKey.CreatedAt, &lastUsedAt, &apiKey.IsActive,
	)

	if err != nil {
		return nil, err
	}

	// Parse permissions JSON
	if err := json.Unmarshal([]byte(permissionsJSON), &apiKey.Permissions); err != nil {
		return nil, fmt.Errorf("failed to unmarshal permissions: %w", err)
	}

	if expiresAt.Valid {
		apiKey.ExpiresAt = &expiresAt.Time
	}

	if lastUsedAt.Valid {
		apiKey.LastUsedAt = &lastUsedAt.Time
	}

	return &apiKey, nil
}

// Health checks the health of the user repository
func (r *UserRepository) Health(ctx context.Context) error {
	query := `SELECT 1 FROM users LIMIT 1`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
