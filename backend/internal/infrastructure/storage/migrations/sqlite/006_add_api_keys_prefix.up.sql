-- Add prefix column to api_keys table for key identification
ALTER TABLE api_keys ADD COLUMN prefix TEXT NOT NULL DEFAULT '';

-- Create index for prefix column for better performance
CREATE INDEX IF NOT EXISTS idx_api_keys_prefix ON api_keys(prefix);
