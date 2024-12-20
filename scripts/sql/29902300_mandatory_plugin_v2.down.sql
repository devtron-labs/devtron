BEGIN;
---
--- Drop all the resource qualifier mappings related to plugin policy version 2.
---
DELETE FROM resource_qualifier_mapping
    WHERE resource_type = 9
    AND resource_id IN (
        SELECT id
        FROM resource_qualifier_mapping_criteria
            WHERE policy_type = 'plugin'
            AND version = 'v1'
     )
;

---
--- Drop all the criteria mappings related to plugin policy version 2.
---
DELETE FROM resource_qualifier_mapping_criteria
    WHERE policy_type = 'plugin'
    AND version = 'v1';

---
--- Drop the global policy history related to plugin policy version 2.
---
DELETE FROM global_policy_history
    WHERE global_policy_id IN (
        SELECT id
        FROM global_policy
            WHERE policy_of = 'PLUGIN'
            AND version = 'V2'
    )
;

---
--- Drop the global policy related to plugin policy version 2.
---
DELETE FROM global_policy
    WHERE policy_of = 'PLUGIN'
    AND version = 'V2';

---
--- Drop the columns policy_type and version from the resource_qualifier_mapping_criteria table.
---
ALTER TABLE "public"."resource_qualifier_mapping_criteria"
    DROP COLUMN IF EXISTS policy_type,
    DROP COLUMN IF EXISTS version;

---
--- Drop the searchable key for CI_PIPELINE_BRANCH_REGEX
---
DELETE FROM devtron_resource_searchable_key
       WHERE name IN ('CI_PIPELINE_BRANCH_REGEX')
         AND is_removed = false;

END;