-- update the container image path for the polling plugin version 1.0.0
UPDATE plugin_pipeline_script
SET container_image_path ='ashexp/polling-plugin:v1.0.0-beta2'
WHERE container_image_path ='quay.io/devtron/poll-container-image:97a996a5-545-16654'
AND deleted = false;