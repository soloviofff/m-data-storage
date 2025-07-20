-- Fix roles table schema to match Role entity
-- Add missing display_name and is_system columns

-- Create new roles table with correct schema
CREATE TABLE roles_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    display_name TEXT NOT NULL,
    description TEXT,
    permissions TEXT NOT NULL, -- JSON array of permissions
    is_system BOOLEAN NOT NULL DEFAULT false,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy existing data from old table
INSERT INTO roles_new (id, name, display_name, description, permissions, is_system, is_active, created_at, updated_at)
SELECT 
    id, 
    name, 
    CASE 
        WHEN name = 'admin' THEN 'Administrator'
        WHEN name = 'user' THEN 'User'
        WHEN name = 'readonly' THEN 'Read Only'
        WHEN name = 'api_access' THEN 'API Access'
        ELSE name
    END as display_name,
    description, 
    permissions,
    CASE 
        WHEN name IN ('admin', 'user', 'readonly', 'api_access') THEN true
        ELSE false
    END as is_system,
    is_active, 
    created_at, 
    updated_at
FROM roles;

-- Drop old table
DROP TABLE roles;

-- Rename new table
ALTER TABLE roles_new RENAME TO roles;

-- Recreate indexes
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_active ON roles(is_active);
CREATE INDEX idx_roles_system ON roles(is_system);

-- Recreate trigger
CREATE TRIGGER roles_updated_at
AFTER UPDATE ON roles
BEGIN
    UPDATE roles SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
