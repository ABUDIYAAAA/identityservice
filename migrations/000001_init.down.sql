DROP INDEX IF EXISTS idx_revoked_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_user_invitations_token_hash;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS revoked_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS user_invitations;
DROP TABLE IF EXISTS users;

DROP TYPE IF EXISTS user_role;