-- Migration 003: Authentication and Authorization Tables
-- This migration creates tables for user management, roles, permissions, sessions, and API keys

-- Create permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create roles table
CREATE TABLE IF NOT EXISTS roles (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    is_system BOOLEAN NOT NULL DEFAULT FALSE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create role_permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id TEXT NOT NULL,
    permission_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'deleted')),
    role_id TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_login_at DATETIME,
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Create user_sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    refresh_token TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ip_address TEXT,
    user_agent TEXT,
    is_revoked BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create api_keys table
CREATE TABLE IF NOT EXISTS api_keys (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    name TEXT NOT NULL,
    key_hash TEXT NOT NULL UNIQUE,
    prefix TEXT NOT NULL,
    permissions TEXT NOT NULL, -- JSON array of permissions
    expires_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_used_at DATETIME,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create security_events table for audit logging
CREATE TABLE IF NOT EXISTS security_events (
    id TEXT PRIMARY KEY,
    user_id TEXT,
    event_type TEXT NOT NULL,
    description TEXT NOT NULL,
    ip_address TEXT,
    user_agent TEXT,
    metadata TEXT, -- JSON object
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    severity TEXT NOT NULL DEFAULT 'low' CHECK (severity IN ('low', 'medium', 'high', 'critical')),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Create password_reset_tokens table
CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL,
    token_hash TEXT NOT NULL UNIQUE,
    expires_at DATETIME NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    used_at DATETIME,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create account_lockouts table
CREATE TABLE IF NOT EXISTS account_lockouts (
    id TEXT PRIMARY KEY,
    user_id TEXT NOT NULL UNIQUE,
    locked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    unlock_at DATETIME NOT NULL,
    reason TEXT,
    locked_by TEXT, -- User ID who locked the account (for admin locks)
    failed_attempts INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (locked_by) REFERENCES users(id) ON DELETE SET NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);
CREATE INDEX IF NOT EXISTS idx_users_last_login ON users(last_login_at);

CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_token_hash ON user_sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_is_revoked ON user_sessions(is_revoked);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(prefix);
CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);

CREATE INDEX IF NOT EXISTS idx_roles_name ON roles(name);
CREATE INDEX IF NOT EXISTS idx_roles_is_system ON roles(is_system);

CREATE INDEX IF NOT EXISTS idx_permissions_name ON permissions(name);
CREATE INDEX IF NOT EXISTS idx_permissions_resource ON permissions(resource);
CREATE INDEX IF NOT EXISTS idx_permissions_action ON permissions(action);
CREATE INDEX IF NOT EXISTS idx_permissions_is_system ON permissions(is_system);

CREATE INDEX IF NOT EXISTS idx_security_events_user_id ON security_events(user_id);
CREATE INDEX IF NOT EXISTS idx_security_events_event_type ON security_events(event_type);
CREATE INDEX IF NOT EXISTS idx_security_events_timestamp ON security_events(timestamp);
CREATE INDEX IF NOT EXISTS idx_security_events_severity ON security_events(severity);

CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_token_hash ON password_reset_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX IF NOT EXISTS idx_password_reset_tokens_is_used ON password_reset_tokens(is_used);

CREATE INDEX IF NOT EXISTS idx_account_lockouts_user_id ON account_lockouts(user_id);
CREATE INDEX IF NOT EXISTS idx_account_lockouts_unlock_at ON account_lockouts(unlock_at);

-- Create triggers for updated_at timestamps
CREATE TRIGGER IF NOT EXISTS update_users_updated_at
    AFTER UPDATE ON users
    FOR EACH ROW
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_roles_updated_at
    AFTER UPDATE ON roles
    FOR EACH ROW
BEGIN
    UPDATE roles SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

CREATE TRIGGER IF NOT EXISTS update_permissions_updated_at
    AFTER UPDATE ON permissions
    FOR EACH ROW
BEGIN
    UPDATE permissions SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert default permissions
INSERT OR IGNORE INTO permissions (id, name, display_name, description, resource, action, is_system) VALUES
-- User management
('users_read', 'users:read', 'Read Users', 'View user information', 'users', 'read', TRUE),
('users_write', 'users:write', 'Write Users', 'Create and update users', 'users', 'write', TRUE),
('users_delete', 'users:delete', 'Delete Users', 'Delete users', 'users', 'delete', TRUE),

-- Role management
('roles_read', 'roles:read', 'Read Roles', 'View role information', 'roles', 'read', TRUE),
('roles_write', 'roles:write', 'Write Roles', 'Create and update roles', 'roles', 'write', TRUE),
('roles_delete', 'roles:delete', 'Delete Roles', 'Delete roles', 'roles', 'delete', TRUE),

-- Instrument management
('instruments_read', 'instruments:read', 'Read Instruments', 'View instrument information', 'instruments', 'read', TRUE),
('instruments_write', 'instruments:write', 'Write Instruments', 'Create and update instruments', 'instruments', 'write', TRUE),
('instruments_delete', 'instruments:delete', 'Delete Instruments', 'Delete instruments', 'instruments', 'delete', TRUE),

-- Data access
('data_read', 'data:read', 'Read Data', 'Access market data', 'data', 'read', TRUE),
('data_write', 'data:write', 'Write Data', 'Store market data', 'data', 'write', TRUE),
('data_delete', 'data:delete', 'Delete Data', 'Delete market data', 'data', 'delete', TRUE),

-- Subscription management
('subscriptions_read', 'subscriptions:read', 'Read Subscriptions', 'View subscriptions', 'subscriptions', 'read', TRUE),
('subscriptions_write', 'subscriptions:write', 'Write Subscriptions', 'Manage subscriptions', 'subscriptions', 'write', TRUE),
('subscriptions_delete', 'subscriptions:delete', 'Delete Subscriptions', 'Delete subscriptions', 'subscriptions', 'delete', TRUE),

-- API key management
('api_keys_read', 'api_keys:read', 'Read API Keys', 'View API keys', 'api_keys', 'read', TRUE),
('api_keys_write', 'api_keys:write', 'Write API Keys', 'Create and update API keys', 'api_keys', 'write', TRUE),
('api_keys_delete', 'api_keys:delete', 'Delete API Keys', 'Delete API keys', 'api_keys', 'delete', TRUE),

-- System administration
('system_read', 'system:read', 'Read System', 'View system information', 'system', 'read', TRUE),
('system_write', 'system:write', 'Write System', 'Modify system settings', 'system', 'write', TRUE),

-- Super admin
('all', '*', 'All Permissions', 'Full system access', '*', '*', TRUE);

-- Insert default roles
INSERT OR IGNORE INTO roles (id, name, display_name, description, is_system) VALUES
('admin', 'admin', 'Administrator', 'Full system access with all permissions', TRUE),
('user', 'user', 'User', 'Standard user with basic data access', TRUE),
('readonly', 'readonly', 'Read Only', 'Read-only access to data and instruments', TRUE),
('api_access', 'api_access', 'API Access', 'Programmatic API access for external systems', TRUE);

-- Assign permissions to roles
INSERT OR IGNORE INTO role_permissions (role_id, permission_id) VALUES
-- Admin gets all permissions
('admin', 'all'),

-- User gets basic permissions
('user', 'instruments_read'),
('user', 'data_read'),
('user', 'subscriptions_read'),
('user', 'subscriptions_write'),
('user', 'api_keys_read'),
('user', 'api_keys_write'),

-- Read-only gets read permissions
('readonly', 'instruments_read'),
('readonly', 'data_read'),
('readonly', 'subscriptions_read'),

-- API access gets data and subscription permissions
('api_access', 'instruments_read'),
('api_access', 'data_read'),
('api_access', 'data_write'),
('api_access', 'subscriptions_read'),
('api_access', 'subscriptions_write');

-- Create default admin user (password: admin123)
-- Password hash for 'admin123' using bcrypt cost 12
INSERT OR IGNORE INTO users (id, username, email, password_hash, first_name, last_name, status, role_id) VALUES
('admin', 'admin', 'admin@m-data-storage.local', '$2a$12$LQv3c1yqBWVHxkd0LHAkCOYz6TtxMQJqhN8/LewdBPj/hL.ckstjC', 'System', 'Administrator', 'active', 'admin');
