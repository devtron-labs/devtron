CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_krr_scan_history_krr_scan_request
    ON krr_scan_history (krr_scan_request);