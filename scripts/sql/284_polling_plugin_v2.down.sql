-- revert the container image path of the polling plugin version 1.0.0
UPDATE plugin_pipeline_script
SET container_image_path ='quay.io/devtron/poll-container-image:97a996a5-545-16654'
WHERE container_image_path ='quay.io/devtron/devtron-plugins:polling-plugin-v1.0.1'
AND deleted = false;