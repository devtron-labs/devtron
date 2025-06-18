BEGIN;

-- Drop table
DROP TABLE IF EXISTS "public"."workflow_config_snapshot";

-- Drop sequence
DROP SEQUENCE IF EXISTS id_seq_workflow_config_snapshot;

COMMIT;
