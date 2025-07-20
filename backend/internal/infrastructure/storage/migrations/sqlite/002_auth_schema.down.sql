-- Migration: Authentication Schema Rollback
-- Description: Drop authentication and authorization tables
-- Created: 2025-07-20

-- Drop triggers
DROP TRIGGER IF EXISTS users_updated_at;
DROP TRIGGER IF EXISTS roles_updated_at;
DROP TRIGGER IF EXISTS api_keys_updated_at;

-- Drop indexes
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_email;
DROP INDEX IF EXISTS idx_users_active;
DROP INDEX IF EXISTS idx_roles_name;
DROP INDEX IF EXISTS idx_roles_active;
DROP INDEX IF EXISTS idx_user_roles_user_id;
DROP INDEX IF EXISTS idx_user_roles_role_id;
DROP INDEX IF EXISTS idx_api_keys_user_id;
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_api_keys_active;
DROP INDEX IF EXISTS idx_api_keys_expires_at;

-- Drop tables (in reverse order due to foreign keys)
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS user_roles;
DROP TABLE IF EXISTS roles;
DROP TABLE IF EXISTS users;
