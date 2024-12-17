BEGIN;

-- Alter Table: global_tag; Add columns deployment_policy and value_constraint_id
ALTER TABLE global_tag
    ADD COLUMN IF NOT EXISTS deployment_policy VARCHAR(50) DEFAULT 'allow' NOT NULL,
    ADD COLUMN IF NOT EXISTS value_constraint_id INT;

COMMIT;