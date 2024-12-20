BEGIN;

-- Alter Table: global_tag; Add columns deployment_policy and value_constraint_id
ALTER TABLE global_tag
    ADD COLUMN IF NOT EXISTS deployment_policy VARCHAR(50) DEFAULT 'allow' NOT NULL,
    ADD COLUMN IF NOT EXISTS value_constraint_id INT;

-- Change column type to TEXT if it is not already TEXT
ALTER TABLE global_tag
ALTER COLUMN mandatory_project_ids_csv TYPE TEXT
    USING mandatory_project_ids_csv::TEXT;

COMMIT;