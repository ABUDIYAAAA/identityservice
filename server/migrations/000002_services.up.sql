-- 1. Services Table
CREATE TABLE IF NOT EXISTS services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    client_id VARCHAR(64) NOT NULL UNIQUE,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    token_version INT NOT NULL DEFAULT 0,
    created_by UUID NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. User-Service Assignments (Which non-admin users manage which service)
CREATE TABLE IF NOT EXISTS user_services (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, service_id)
);

-- 3. Service Secrets (Multiple active secrets per service)
CREATE TABLE IF NOT EXISTS service_secrets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL DEFAULT 'Default Secret',
    secret_prefix VARCHAR(10) NOT NULL,
    secret_hash VARCHAR(64) NOT NULL UNIQUE,
    last_used_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 4. Service Permissions (Direct Link Matrix between Services)
CREATE TABLE IF NOT EXISTS service_permissions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE, -- Calling service
    target_service_id UUID NOT NULL REFERENCES services(id) ON DELETE CASCADE, -- Target/verifying service
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(source_service_id, target_service_id)
);

-- Indexes
CREATE INDEX idx_services_client_id ON services(client_id);
CREATE INDEX idx_service_secrets_service_id ON service_secrets(service_id);
CREATE INDEX idx_service_secrets_hash ON service_secrets(secret_hash);
CREATE INDEX idx_user_services_user_id ON user_services(user_id);
CREATE INDEX idx_service_permissions_lookup ON service_permissions(source_service_id, target_service_id);