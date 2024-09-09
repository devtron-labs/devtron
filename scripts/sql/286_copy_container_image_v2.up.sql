

INSERT INTO "plugin_parent_metadata" ("id", "name", "identifier", "description", "type", "icon", "deleted", "created_on", "created_by", "updated_on", "updated_by")
SELECT nextval('id_seq_plugin_parent_metadata'), 'Copy container image','copy-container-image', 'Copy container images from the source repository to a desired repository','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/ic-plugin-copy-container-image.png','f', 'now()', 1, 'now()', 1
    WHERE NOT EXISTS (
    SELECT 1
    FROM plugin_parent_metadata
    WHERE identifier='copy-container-image'
    AND deleted = false
);

-- update the plugin_metadata with the plugin_parent_metadata_id
UPDATE plugin_metadata
SET plugin_parent_metadata_id = (
    SELECT id
    FROM plugin_parent_metadata
    WHERE identifier='copy-container-image'
      AND deleted = false
)
WHERE name='Copy container image'
  AND (
        plugin_parent_metadata_id IS NULL
        OR plugin_parent_metadata_id = 0
    )
  AND deleted = false;
