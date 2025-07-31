BEGIN;

-- Drop table
DROP TABLE IF EXISTS "public"."bulk_edit_config";

-- Drop sequence
DROP SEQUENCE IF EXISTS id_seq_bulk_edit_config;

COMMIT;