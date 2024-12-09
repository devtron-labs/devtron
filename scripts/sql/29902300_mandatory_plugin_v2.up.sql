BEGIN;
---
--- add policy_type and version columns to resource_qualifier_mapping_criteria table
---
ALTER TABLE "public"."resource_qualifier_mapping_criteria"
    ADD COLUMN IF NOT EXISTS policy_type varchar(50) NOT NULL DEFAULT 'lock-configuration',
    ADD COLUMN IF NOT EXISTS version varchar(50) NOT NULL DEFAULT 'v1';

---
--- drop default constraint for policy_type column, as it is not necessary
---
ALTER TABLE "public"."resource_qualifier_mapping_criteria"
    ALTER COLUMN policy_type DROP DEFAULT;

---
--- add searchable key for CI_PIPELINE_BRANCH_REGEX
---
INSERT INTO devtron_resource_searchable_key(name, is_removed, created_on, created_by, updated_on, updated_by)
SELECT 'CI_PIPELINE_BRANCH_REGEX', false, now(), 1, now(), 1
    WHERE NOT EXISTS (
    SELECT 1
    FROM devtron_resource_searchable_key
    WHERE name='CI_PIPELINE_BRANCH_REGEX'
    AND is_removed = false)
;
END;