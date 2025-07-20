-- Remove prefix column and index from api_keys table
DROP INDEX IF EXISTS idx_api_keys_prefix;

-- Note: SQLite doesn't support DROP COLUMN directly
-- This would require table recreation in a real rollback scenario
-- For now, we'll leave the column as it doesn't break functionality
