INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Pull images from container repository','Polls a container repository and pulls images stored in the repository which can be used for deployment.','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/plugin-poll-container-registry.png','f', 'now()', 1, 'now()', 1);

INSERT INTO "plugin_tag" ("id", "name", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag'), 'Image source','f', 'now()',1, 'now()', 1);

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='Image source') , (SELECT id FROM plugin_metadata WHERE name='Pull images from container repository'),'now()', 1, 'now()', 1);

INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name='Pull images from container repository'),0,'now()', 1, 'now()', 1);

INSERT INTO "plugin_pipeline_script" ("id","type","mount_directory_from_host","container_image_path","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),'CONTAINER_IMAGE','t','quay.io/devtron/poll-container-image:f7309681-545-16560','f','now()',1,'now()',1);

INSERT INTO "script_path_arg_port_mapping" ("id", "type_of_mapping", "file_path_on_disk","file_path_on_container","script_id","deleted","created_on","created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_script_path_arg_port_mapping'),'FILE_PATH','/polling-plugin' ,'/output',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Pull images from container repository'),'Step 1','Step 1 - Polling from Container Registry','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Pull images from container repository' and ps."index"=1 and ps.deleted=false), 'REPOSITORY','STRING','Provide the repository name for polling.',true,false,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Pull images from container repository' and ps."index"=1 and ps.deleted=false), 'ACCESS_KEY','STRING','Aws Ecr Access Key',false,true,'INPUT','GLOBAL',1 ,'ACCESS_KEY','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Pull images from container repository' and ps."index"=1 and ps.deleted=false), 'SECRET_KEY','STRING','AWS ECR Secret Key',false,true,'INPUT','GLOBAL',1 ,'SECRET_KEY','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Pull images from container repository' and ps."index"=1 and ps.deleted=false), 'LAST_FETCHED_TIME','STRING','CR Last fetched time',false,true,'INPUT','GLOBAL',1 ,'LAST_FETCHED_TIME','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Pull images from container repository' and ps."index"=1 and ps.deleted=false), 'DOCKER_REGISTRY_URL','STRING','CR Docker registry Url',false,true,'INPUT','GLOBAL',1 ,'DOCKER_REGISTRY_URL','f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Pull images from container repository' and ps."index"=1 and ps.deleted=false), 'AWS_REGION','STRING','CR Aws region',false,true,'INPUT','GLOBAL',1 ,'AWS_REGION','f','now()', 1, 'now()', 1);