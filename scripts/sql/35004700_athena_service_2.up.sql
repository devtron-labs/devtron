-- 2_alter_runbook_approval_up.sql
-- Added unique constraint on (runbook_id, cluster_id)
-- Make sure there are no duplicate entries before adding the constraint
ALTER TABLE runbook_approval_status
    ADD CONSTRAINT uq_runbook_cluster
        UNIQUE (runbook_id, cluster_id);

ALTER TABLE public.master_audit_logs ALTER COLUMN "action" TYPE varchar(255);