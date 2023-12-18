update custom_tag set enabled=true where active=true;
update plugin_pipeline_script set container_image_path ='quay.io/devtron/copy-container-images:48bdcbb4-567-19597'
                              where container_image_path ='quay.io/devtron/copy-container-images:7285439d-567-19519';