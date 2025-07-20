-- Migration: Fix Users Table Schema (Rollback)
-- Description: Rollback users table to original schema
-- Created: 2025-07-20

-- Recreate original users table
CREATE TABLE users_old (
    id TEXT PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    email TEXT NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Copy data back, converting status to is_active
INSERT INTO users_old (id, username, email, password_hash, created_at, updated_at, is_active)
SELECT id, username, email, password_hash, created_at, updated_at, 
       CASE WHEN status = 'active' THEN 1 ELSE 0 END as is_active
FROM users;

-- Drop new table and rename old table
DROP TABLE users;
ALTER TABLE users_old RENAME TO users;

-- Recreate original indexes
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_active ON users(is_active);

-- Recreate original trigger
CREATE TRIGGER IF NOT EXISTS users_updated_at
AFTER UPDATE ON users
BEGIN
    UPDATE users SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
