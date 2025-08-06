BEGIN;

-- Drop cd_workflow_status_latest table
DROP TABLE IF EXISTS "public"."cd_workflow_status_latest";

-- Drop sequence for cd_workflow_status_latest
DROP SEQUENCE IF EXISTS id_seq_cd_workflow_status_latest;

-- Drop ci_workflow_status_latest table
DROP TABLE IF EXISTS "public"."ci_workflow_status_latest";

-- Drop sequence for ci_workflow_status_latest
DROP SEQUENCE IF EXISTS id_seq_ci_workflow_status_latest;

COMMIT;
