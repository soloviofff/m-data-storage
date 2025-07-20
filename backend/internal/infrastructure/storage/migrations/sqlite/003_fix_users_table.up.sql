-- Migration: Fix Users Table Schema
-- Description: Update users table to include authentication fields
-- Created: 2025-07-20

-- Enable foreign keys
PRAGMA foreign_keys = ON;

-- Recreate users table with authentication fields
-- Step 1: Create new users table with all required fields
CREATE TABLE users_new (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    first_name TEXT,
    last_name TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive', 'suspended', 'deleted')),
    role_id TEXT,
    last_login_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (role_id) REFERENCES roles(id)
);

-- Step 2: Copy data from old table to new table, converting is_active to status
INSERT INTO users_new (id, username, email, password_hash, created_at, updated_at, status)
SELECT id, username, email, password_hash, created_at, updated_at, 
       CASE WHEN is_active = 1 THEN 'active' ELSE 'inactive' END as status
FROM users;

-- Step 3: Drop old table and rename new table
DROP TABLE users;
ALTER TABLE users_new RENAME TO users;

-- Recreate indexes for users table
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);

-- Recreate trigger for automatic timestamp updates
CREATE TRIGGER IF NOT EXISTS users_updated_at
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
