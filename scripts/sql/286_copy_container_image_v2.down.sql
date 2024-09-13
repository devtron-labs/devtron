-- update the is_latest flag to true, for the plugin version 1.0.0
UPDATE plugin_metadata
SET is_latest = true
WHERE id = (
    SELECT id
    FROM plugin_metadata
    WHERE name='Copy container image'
      AND plugin_version='1.0.0'
      AND is_latest = false
      AND deleted = false
);

DO
$$
DECLARE
temprow record;
BEGIN
    -- select the plugin_metadata_id, plugin_step_id, plugin_pipeline_script_id
SELECT plugin_metadata.id AS plugin_metadata_id,
       plugin_step.id AS plugin_step_id,
       plugin_step.script_id AS plugin_pipeline_script_id
INTO temprow
FROM plugin_metadata
         LEFT JOIN plugin_step ON plugin_metadata.id = plugin_step.plugin_id
WHERE plugin_metadata.plugin_version='2.0.0'
  AND plugin_metadata.name='Copy container image';

-- down script for the new plugin version 2.0.0
DELETE FROM plugin_step_variable
WHERE plugin_step_id = temprow.plugin_step_id;
DELETE FROM script_path_arg_port_mapping
WHERE script_id = temprow.plugin_pipeline_script_id;
DELETE FROM plugin_step
WHERE plugin_id = temprow.plugin_metadata_id;
DELETE FROM plugin_stage_mapping
WHERE plugin_id = temprow.plugin_metadata_id;
DELETE FROM pipeline_stage_step_variable
WHERE pipeline_stage_step_id IN (
    SELECT id
    FROM pipeline_stage_step
    WHERE ref_plugin_id = temprow.plugin_metadata_id
);
DELETE FROM pipeline_stage_step
WHERE ref_plugin_id = temprow.plugin_metadata_id;
DELETE from plugin_pipeline_script
WHERE id = temprow.plugin_pipeline_script_id;
DELETE FROM plugin_tag_relation
WHERE plugin_id = temprow.plugin_metadata_id;
DELETE FROM plugin_metadata
WHERE id = temprow.plugin_metadata_id;

END;
$$;