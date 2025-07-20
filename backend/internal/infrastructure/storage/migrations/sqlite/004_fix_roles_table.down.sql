-- Rollback roles table schema fix
-- Remove display_name and is_system columns

-- Create old roles table schema
CREATE TABLE roles_old (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT NOT NULL, -- JSON array of permissions
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy data back to old schema
INSERT INTO roles_old (id, name, description, permissions, is_active, created_at, updated_at)
SELECT id, name, description, permissions, is_active, created_at, updated_at
FROM roles;

-- Drop new table
DROP TABLE roles;

-- Rename old table
ALTER TABLE roles_old RENAME TO roles;

-- Recreate indexes
CREATE INDEX idx_roles_name ON roles(name);
CREATE INDEX idx_roles_active ON roles(is_active);

-- Recreate trigger
CREATE TRIGGER roles_updated_at
AFTER UPDATE ON roles
BEGIN
    UPDATE roles SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
