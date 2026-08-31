-- Audit Logs Table
CREATE TABLE IF NOT EXISTS audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    action_type VARCHAR(100) NOT NULL,
    actor_type VARCHAR(20) NOT NULL DEFAULT 'user', -- 'user', 'service', 'system'
    actor_id UUID,
    service_id UUID REFERENCES services(id) ON DELETE SET NULL, -- Native service scope support
    before_state JSONB,
    after_state JSONB,
    status VARCHAR(20) NOT NULL DEFAULT 'success', -- 'success', 'failure'
    error_message TEXT,
    ip_address VARCHAR(45),
    user_agent TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_audit_logs_action ON audit_logs(action_type);
CREATE INDEX idx_audit_logs_actor ON audit_logs(actor_id);
CREATE INDEX idx_audit_logs_service ON audit_logs(service_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);

