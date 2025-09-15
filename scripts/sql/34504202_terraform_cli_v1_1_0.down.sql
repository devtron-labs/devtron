BEGIN;

-- Delete plugin step variable (for step index=1)
DELETE FROM plugin_step_variable 
WHERE plugin_step_id = (
    SELECT ps.id 
    FROM plugin_metadata p
    INNER JOIN plugin_step ps ON ps.plugin_id = p.id
    WHERE p.plugin_version = '1.0.1'
      AND p.name = 'Terraform CLI v1.0.0'
      AND p.deleted = false
      AND ps."index" = 1
      AND ps.deleted = false
);

-- Delete plugin steps
DELETE FROM plugin_step 
WHERE plugin_id = (
    SELECT id 
    FROM plugin_metadata 
    WHERE plugin_version = '1.0.1' 
      AND name = 'Terraform CLI v1.0.0' 
      AND deleted = false
);

-- Delete pipeline script
DELETE FROM plugin_pipeline_script 
WHERE id = (
    SELECT script_id 
    FROM plugin_step 
    WHERE plugin_id = (
        SELECT id 
        FROM plugin_metadata 
        WHERE plugin_version = '1.0.1' 
          AND name = 'Terraform CLI v1.0.0' 
          AND deleted = false
    )
);

-- Delete stage mappings
DELETE FROM plugin_stage_mapping 
WHERE plugin_id = (
    SELECT id 
    FROM plugin_metadata 
    WHERE plugin_version = '1.0.1' 
      AND name = 'Terraform CLI v1.0.0' 
      AND deleted = false
);

-- Delete plugin metadata entry
DELETE FROM plugin_metadata 
WHERE plugin_version = '1.0.1' 
  AND name = 'Terraform CLI v1.0.0' 
  AND deleted = false;

-- Mark previous version as latest
UPDATE plugin_metadata 
SET is_latest = true 
WHERE id = (
    SELECT id 
    FROM plugin_metadata 
    WHERE name = 'Terraform CLI v1.0.0'
      AND is_latest = false
      AND plugin_version = '1.0.0'
);

COMMIT;