INSERT INTO "plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_metadata'), 'Jenkins','','PRESET','https://raw.githubusercontent.com/devtron-labs/devtron/main/assets/semgrep.png','f', 'now()', 1, 'now()', 1);



INSERT INTO "plugin_tag" ("id", "name", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag'), 'CI task','f', 'now()',1, 'now()', 1);


INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_tag_relation'),(SELECT id FROM plugin_tag WHERE name='CI task') , (SELECT id FROM plugin_metadata WHERE name='Jenkins'),'now()', 1, 'now()', 1);


INSERT INTO "plugin_stage_mapping" ("plugin_id","stage_type","created_on", "created_by", "updated_on", "updated_by") VALUES ((SELECT id FROM plugin_metadata WHERE name='Jenkins'),0,'now()', 1, 'now()', 1);


INSERT INTO "plugin_pipeline_script" ("id","type","mount_directory_from_host","container_image_path","deleted","created_on", "created_by", "updated_on", "updated_by") VALUES (nextval('id_seq_plugin_pipeline_script'),'CONTAINER_IMAGE','t','quay.io/devtron/test:jenkins-plugin','f','now()',1,'now()',1);

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step'), (SELECT id FROM plugin_metadata WHERE name='Jenkins'),'Step 1','Step 1 - Triggering Jenkins Job','1','INLINE',(SELECT last_value FROM id_seq_plugin_pipeline_script),'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jenkins' and ps."index"=1 and ps.deleted=false), 'URL','STRING','Jenkins account url',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jenkins' and ps."index"=1 and ps.deleted=false), 'USERNAME','STRING','Jenkins account username',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jenkins' and ps."index"=1 and ps.deleted=false), 'PASSWORD','STRING','Jenkins account password/token',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jenkins' and ps."index"=1 and ps.deleted=false), 'JOB_NAME','STRING','Jenkins job name',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jenkins' and ps."index"=1 and ps.deleted=false), 'JOB_TRIGGER_PARAMS','STRING','Jenkins job parameters. These parameters are same as jenkins job trigger params. Provide in format {"key1":"val1","key2":"val2"}',true,true,'INPUT','NEW',1 ,'f','now()', 1, 'now()', 1);

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value","variable_type", "value_type", "variable_step_index",reference_variable_name, "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_step_variable'), (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Jenkins' and ps."index"=1 and ps.deleted=false), 'GIT_MATERIAL_REQUEST','STRING','',false,true,'INPUT','GLOBAL',1 ,'GIT_MATERIAL_REQUEST','f','now()', 1, 'now()', 1);