-- File: 2_alter_runbook_approval_down.sql
-- Drop the unique constraint
ALTER TABLE runbook_approval_status
DROP CONSTRAINT IF EXISTS uq_runbook_cluster;