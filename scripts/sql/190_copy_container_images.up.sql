
-- copy container images plugin migration script start

INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Copy container image','Copy container images from the source repository to a desired repository','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/ic-plugin-copy-container-image.png','f', 'now()', 1, 'now()', 1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='Image source') , (SELECT id FROM plugin_metadata WHERE name='Copy container image'),'now()', 1, 'now()', 1);

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Copy container image'),0,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id","type","mount_directory_from_host","container_image_path","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),'CONTAINER_IMAGE','t','quay.io/devtron/copy-container-images:7285439d-567-19519','f','now()',1,'now()',1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Copy container image'),'Step 1','Step 1 - Copy container images','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and ps."index"=1 and ps.deleted=false), 'DESTINATION_INFO','STRING',
        'In case of CI, build image will be copied to registry and repository provided in DESTINATION_INFO. In case of PRE-CD/POST-CD, Image used to trigger stage will be copied in DESTINATION_INFO
        Format:
            <registry1> | <repo1>,<repo2>', true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and ps."index"=1 and ps.deleted=false), 'DOCKER_IMAGE','STRING','',false,true,'INPUT','GLOBAL',1 ,'DOCKER_IMAGE','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and ps."index"=1 and ps.deleted=false), 'REGISTRY_DESTINATION_IMAGE_MAP','STRING','map of registry name and images needed to be copied in that images',false,true,'INPUT','GLOBAL',1 ,'REGISTRY_DESTINATION_IMAGE_MAP','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and ps."index"=1 and ps.deleted=false), 'REGISTRY_CREDENTIALS','STRING','',false,true,'INPUT','GLOBAL',1 ,'REGISTRY_CREDENTIALS','f','now()', 1, 'now()', 1);

-- copy container images plugin migration script ends

-- requiered db changes for above scipt

ALTER TABLE custom_tag ADD COLUMN enabled boolean default false;
ALTER TABLE ci_artifact ADD COLUMN credentials_source_type VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN credentials_source_value VARCHAR(50);
ALTER TABLE ci_artifact ADD COLUMN  component_id integer;

ALTER TABLE ci_workflow ADD COLUMN image_path_reservation_ids integer[];

UPDATE ci_workflow set image_path_reservation_ids=ARRAY["image_path_reservation_id"] where image_path_reservation_id is not NULL;

ALTER TABLE cd_workflow_runner ADD COLUMN image_path_reservation_ids integer[];

ALTER TABLE image_path_reservation  DROP CONSTRAINT image_path_reservation_custom_tag_id_fkey;