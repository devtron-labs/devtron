BEGIN;

-- Drop Table: infra_config_trigger_history
DROP TABLE IF EXISTS "public"."infra_config_trigger_history";

-- Drop Sequence for infra_config_trigger_history
DROP SEQUENCE IF EXISTS "public"."id_seq_infra_config_trigger_history";

--hard deleting the entries
DELETE  FROM "public"."infra_profile_configuration"
WHERE key = 8 or key =9;

END;