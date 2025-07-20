-- Drop user_sessions table and related objects
DROP TRIGGER IF EXISTS update_user_sessions_last_used_at;
DROP INDEX IF EXISTS idx_user_sessions_is_revoked;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;
DROP INDEX IF EXISTS idx_user_sessions_refresh_token;
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP TABLE IF EXISTS user_sessions;
