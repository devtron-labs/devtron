## Steps to create a new Plugin - 

1. **Enter the plugin's metadata in table `plugin_metadata`**

INSERT INTO
"plugin_metadata" ("id", "name", "description","type","icon","deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES 
(nextval('id_seq_plugin_metadata'), 'name_of_plugin','description_of_plugin','SHARED','link_to_icon','f', 'now()', 'user_id', 'now()', 'user_id');

2. **Add tags for plugin if not already present in the table `plugin_tag`(we will use id of this table for connecting with plugin, if a tag is already present note its id).**

To add a new tag - 

INSERT INTO "plugin_tag" ("id", "name", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_tag'), 'name_of_tag','f', 'now()', 'user_id', 'now()', 'user_id');

3. **Update plugin & tag relation in the table `plugin_tag_relation`**

INSERT INTO "plugin_tag_relation" ("id", "tag_id", "plugin_id", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_tag_relation'), 'id_from_table-plugin_tag','id_from_table-plugin_metadata','now()', 'user_id', 'now()', 'user_id');

Note - there can be multiple tags associated with a plugin.

4. **Create scripts for plugin steps (for every custom type step in your plugin, you need to create script before adding that step)**

To create script- 

INSERT INTO "plugin_pipeline_script" ("id", "script", "store_script_at","mount_code_to_container,"mount_code_to_container_path", "mount_directory_from_host","type","deleted","created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_pipeline_script'), 'script','mount_code_at-field-on-UI',"true/false","path_if_yes_in_previous_field","true/false",'SHELL/CONTAINER_IMAGE','f','now()', 'user_id', 'now()', 'user_id');

After script is created, if there are mappings (filePath mapping in mount_directory_from_host option, command args, port mapping) available in the script add them in table `script_path_arg_port_mapping` -

INSERT INTO "script_path_arg_port_mapping" ("id", "type_of_mapping", "file_path_on_disk","file_path_on_container,"command", "args","port_on_local","port_on_container","script_id","deleted","created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_script_path_arg_port_mapping'), 'FILE_PATH/DOCKER_ARG/PORT','file_path_mapping_entry',"file_path_mapping_entry","command","array_of_args",'port_on_local',''port_on_container','id-from-script-table','f','now()', 'user_id', 'now()', 'user_id');

Note - For multiple mappings (file or port) in a script, separate entry is needed for each mapping

5. **Create plugin steps**

To create a step - 

INSERT INTO "plugin_step" ("id", "plugin_id","name","description","index","step_type","script_id","ref_plugin_id","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step'), 'id-from-plugin_metadata','name_of_step','description_of_step','index_of_step-start-from-1','INLINE/REF_PLUGIN','id_from_script_table','id-of-ref-plugin-get-from-plugin_metadata','f','now()', 'user_id', 'now()', 'user_id');

6. **Create variables present in steps**

To create variable - 

INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "previous_step_index","default_value", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_variable'), 'id-from-plugin_step','name_of_variable','STRING/BOOL/DATE/NUMBER','description_of_variable','true/false','true/false','INPUT/OUTPUT','NEW/FROM_PREVIOUS_STEP/GLOBAL','index_of_step_of_ref_variable','default_value-nullable','index_of_step_variable_is_present_in','f','now()', 'user_id', 'now()', 'user_id');

7. **Create conditions on variables**

To create condition -

INSERT INTO "plugin_step_condition" ("id", "plugin_step_id","condition_variable_id","condition_type", "conditional_operator","conditional_value","deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
(nextval('id_seq_plugin_step_condition'), 'id-from-plugin_step',"id-of-variable-on-which-condition-is-applied",'SKIP/TRIGGER/SUCCESS/FAILURE','conditional_operator','conditional_value','f','now()', 'user_id', 'now()', 'user_id');

