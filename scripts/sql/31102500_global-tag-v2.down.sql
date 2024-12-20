BEGIN;

-- Revert column type change
ALTER TABLE global_tag
ALTER COLUMN mandatory_project_ids_csv TYPE VARCHAR(100)
    USING mandatory_project_ids_csv::VARCHAR(100);

-- Drop columns deployment_policy and value_constraint_id
ALTER TABLE global_tag
DROP COLUMN IF EXISTS deployment_policy,
    DROP COLUMN IF EXISTS value_constraint_id;

COMMIT;