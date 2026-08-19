DROP INDEX IF EXISTS idx_service_permissions_lookup;
DROP INDEX IF EXISTS idx_user_services_user_id;
DROP INDEX IF EXISTS idx_service_secrets_hash;
DROP INDEX IF EXISTS idx_service_secrets_service_id;
DROP INDEX IF EXISTS idx_services_client_id;

DROP TABLE IF EXISTS service_permissions;
DROP TABLE IF EXISTS service_secrets;
DROP TABLE IF EXISTS user_services;
DROP TABLE IF EXISTS services;