BEGIN;

-- Alter Table: global_tag; Drop columns deployment_policy and value_constraint_id if they exist
ALTER TABLE global_tag
DROP COLUMN IF EXISTS deployment_policy,
    DROP COLUMN IF EXISTS value_constraint_id;

COMMIT;