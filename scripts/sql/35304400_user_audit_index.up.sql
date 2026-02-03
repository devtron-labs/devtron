CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_user_audit_user_id_id_desc
    ON user_audit (user_id, id DESC);