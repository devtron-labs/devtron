INSERT INTO "plugin_parent_metadata" ("id", "name","identifier", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_parent_metadata'), 'Copy container image','copy-container-image','Copy container images from the source repository to a desired repository','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/ic-plugin-copy-container-image.png','f', 'now()', 1, 'now()', 1)
ON CONFLICT (identifier) DO NOTHING;

update plugin_metadata set plugin_parent_metadata_id = (select id from plugin_parent_metadata where identifier='copy-container-image') where name='Copy container image';

UPDATE plugin_metadata SET is_latest = false WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'Copy container image' and is_latest= true);


INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'Copy container image','Copy container images from the source repository to a desired repository','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='copy-container-image'),'2.0.0', false, true);


INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='Image source') , (SELECT id FROM plugin_metadata WHERE name='Copy container image' and plugin_version='1.0.0'),'now()', 1, 'now()', 1);


INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='2.0.0' and name='Copy container image' and deleted= false),0,'now()', 1, 'now()', 1);


INSERT INTO "plugin_pipeline_script" ("id","type","mount_directory_from_host","container_image_path","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),'CONTAINER_IMAGE','t','quay.io/devtron/test:b7b9ba5c-10-37','f','now()',1,'now()',1);

INSERT INTO "script_path_arg_port_mapping" ("id", "type_of_mapping", "file_path_on_disk","file_path_on_container","script_id","deleted","created_on","created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_script_path_arg_port_mapping'),'FILE_PATH','/copy-container-image' ,'/output',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'),(SELECT id FROM plugin_metadata WHERE plugin_version='2.0.0' and name='Copy container image' and deleted= false),'Step 1','Step 1 - Copy container images','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='2.0.0' and p.name='Copy container image' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'IS_IMAGE_SCAN','BOOL','Flag to identify if this plugin used for image scan',false,true,'INPUT','NEW','true',1 ,'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and p.plugin_version='2.0.0' and ps."index"=1 and ps.deleted=false), 'DESTINATION_INFO','STRING',
        'In case of CI, build image will be copied to registry and repository provided in DESTINATION_INFO. In case of PRE-CD/POST-CD, Image used to trigger stage will be copied in DESTINATION_INFO
        Format:
            <registry1> | <repo1>,<repo2>', true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and p.plugin_version='2.0.0'  and ps."index"=1 and ps.deleted=false), 'DOCKER_IMAGE','STRING','',false,true,'INPUT','GLOBAL',1 ,'DOCKER_IMAGE','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and p.plugin_version='2.0.0'  and ps."index"=1 and ps.deleted=false), 'REGISTRY_DESTINATION_IMAGE_MAP','STRING','map of registry name and images needed to be copied in that images',false,true,'INPUT','GLOBAL',1 ,'REGISTRY_DESTINATION_IMAGE_MAP','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and p.plugin_version='2.0.0'  and ps."index"=1 and ps.deleted=false), 'REGISTRY_CREDENTIALS','STRING','',false,true,'INPUT','GLOBAL',1 ,'REGISTRY_CREDENTIALS','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and p.plugin_version='2.0.0'  and ps."index"=1 and ps.deleted=false), 'DOCKER_IMAGE_TAG','STRING','',false,true,'INPUT','GLOBAL',1 ,'DOCKER_TAG','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Copy container image' and p.plugin_version='2.0.0'  and ps."index"=1 and ps.deleted=false), 'DOCKER_IMAGE_TAG_OVERRIDE','STRING','',true,true,'INPUT','FROM_PREVIOUS_STEP',1 ,'DOCKER_IMAGE_TAG_OVERRIDE','f','now()', 1, 'now()', 1);
