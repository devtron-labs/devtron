-- update the container image path for the polling plugin version 1.0.0
UPDATE plugin_pipeline_script
SET container_image_path ='quay.io/devtron/devtron-plugins:polling-plugin-v1.0.1'
WHERE container_image_path ='quay.io/devtron/poll-container-image:97a996a5-545-16654'
AND deleted = false;

-- create plugin_parent_metadata for the polling plugin, if not exists
INSERT INTO "plugin_parent_metadata" ("id", "name", "identifier", "description", "type", "icon", "deleted", "created_on", "created_by", "updated_on", "updated_by")
SELECT nextval('id_seq_plugin_parent_metadata'), 'Pull images from container repository','pull-images-from-container-repository','Polls a container repository and pulls images stored in the repository which can be used for deployment.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/plugin-poll-container-registry.png','f', 'now()', 1, 'now()', 1
    WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_parent_metadata
    WHERE identifier='pull-images-from-container-repository'
    AND deleted = false
);

-- update the plugin_metadata with the plugin_parent_metadata_id
UPDATE plugin_metadata
SET plugin_parent_metadata_id = (
    SELECT id
    FROM plugin_parent_metadata
    WHERE identifier='pull-images-from-container-repository'
      AND deleted = false
)
WHERE name='Pull images from container repository'
  AND (
    plugin_parent_metadata_id IS NULL
        OR plugin_parent_metadata_id = 0
    )
  AND deleted = false;