-- Drop indexes
DROP INDEX IF EXISTS idx_email_verifications_user_id;
DROP INDEX IF EXISTS idx_email_verifications_token;
DROP INDEX IF EXISTS idx_password_resets_user_id;
DROP INDEX IF EXISTS idx_password_resets_token;

-- Drop tables
DROP TABLE IF EXISTS email_verifications;
DROP TABLE IF EXISTS password_resets; 