INSERT INTO "plugin_parent_metadata" ("id", "name","identifier", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_parent_metadata'), 'Copy container image','copy-image-plugin','Copy container images from the source repository to a desired repository','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/ic-plugin-copy-container-image.png','f', 'now()', 1, 'now()', 1);


UPDATE plugin_metadata SET is_latest = false WHERE id = (SELECT id FROM plugin_metadata WHERE name= 'Copy container image' and is_latest= true);


INSERT INTO "plugin_metadata" ("id", "name", "description","deleted", "created_on", "created_by", "updated_on", "updated_by","plugin_parent_metadata_id","plugin_version","is_deprecated","is_latest")
VALUES (nextval('id_seq_plugin_metadata'), 'Copy container image','Copy container images from the source repository to a desired repository','f', 'now()', 1, 'now()', 1, (SELECT id FROM plugin_parent_metadata WHERE identifier='Copy container image'),'2.0.0', false, true);


INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='Image source') , (SELECT id FROM plugin_metadata WHERE name='Copy container image' and plugin_version='1.0.0'),'now()', 1, 'now()', 1);


INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE plugin_version='2.0.0' and name='Copy container image' and deleted= false),3,'now()', 1, 'now()', 1);


INSERT INTO "plugin_pipeline_script" ("id","type","mount_directory_from_host","container_image_path","deleted","created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_pipeline_script'),'CONTAINER_IMAGE','t','','f','now()',1,'now()',1);

INSERT INTO "script_path_arg_port_mapping" ("id", "type_of_mapping", "file_path_on_disk","file_path_on_container","script_id","deleted","created_on","created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_script_path_arg_port_mapping'),'FILE_PATH','/polling-plugin' ,'/output',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'),(SELECT id FROM plugin_metadata WHERE plugin_version='2.0.0' and name='Copy container image' and deleted= false),'Step 1','Step 1 - Copy container images','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='IMAGE SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'IS_IMAGE_SCAN','BOOL','Flag to identify if this plugin used for image scan',false,true,'INPUT','NEW','true',1 ,'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='IMAGE SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'SCAN_VIA_V2','BOOL','Boolean flag to identify which scanning method will be used to scan an image',true,false,'INPUT','NEW','false',1 ,'f','now()', 1, 'now()', 1);


INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='IMAGE SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'SCAN_RETRY_COUNT','NUMBER','image scanner retry count',true,false,'INPUT','NEW',3,'1' ,'f','now()', 1, 'now()', 1);








INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='IMAGE SCAN' and p.deleted=false and ps."index"=1 and ps.deleted=false), 'SCAN_RETRY_INTERVAL','NUMBER','image scanner endpoint',true,false,'INPUT','NEW',2,1 ,'f','now()', 1, 'now()', 1);
