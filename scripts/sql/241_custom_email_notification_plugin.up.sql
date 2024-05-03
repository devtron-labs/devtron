-- Insert Plugin Metadata
INSERT INTO "plugin_metadata" ("id", "name", "description", "type", "icon", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Custom Email Notifier v1.0.0', 'Send email notifications', 'PRESET', 'https://t4.ftcdn.net/jpg/04/76/40/09/240_F_476400965_iYkCpMeqvQwaICySZBgyP5IErlyAgeZu.jpg', 'f', 'now()', 1, 'now()', 1);

-- Insert Plugin Stage Mapping
INSERT INTO "plugin_stage_mapping" ("plugin_id", "stage_type", "created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name = 'Custom Email Notifier v1.0.0'), 0, 'now()', 1, 'now()', 1);

-- Insert Plugin Script
INSERT INTO "plugin_pipeline_script" ("id", "script", "type", "deleted", "created_on", "created_by", "updated_giton", "updated_by")
VALUES (
           nextval('id_seq_plugin_pipeline_script'),
           E'#!/bin/sh

    # URL and token from environment variables
    echo "------------STARTING PLUGIN CUSTOM EMAIL NOTIFIER------------"
    echo "CONFIG_TYPE: $CONFIG_TYPE"
    echo "CONFIG_NAME: $CONFIG_NAME"
    echo "EMAIL_IDS: $EMAIL_IDS"
    url = $NOTIFICATION_URL
    token = $NOTIFICATION_TOKEN
    echo "NOTIFICATION_URL: $NOTIFICATION_URL"
    echo "NOTIFICATION_TOKEN: $NOTIFICATION_TOKEN"
    echo "{"configType": \'"${CONFIG_TYPE}"\', "configName": \'"${CONFIG_NAME}"\', "emailIds": \'"$EMAIL_IDS"\'}"
    echo \'{"configType":\'${CONFIG_TYPE}\',"configName":\'${CONFIG_NAME}\',"emailIds":\'${EMAIL_IDS}\'}\'
    # Make the API call
    curl -X POST https://devtron-14.devtron.info/orchestrator/scoop/intercept-event/notify?clusterId=1 -H "token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3MTQ3NDU5NzUsImp0aSI6IjI4Y2Y2YWQzLTY0OGYtNDhjNS04MjIzLTFhZGY3MzExZDQ3YiIsImlhdCI6MTcxNDY1OTU3NSwiaXNzIjoiYXJnb2NkIiwibmJmIjoxNzE0NjU5NTc1LCJzdWIiOiJhZG1pbiJ9.xIWS9sRDYBqbh0GTWtt9nHlSD3GCzyJNjrNaOc1W_W8" -H "Content-Type: application/json" -d ''{"configType": "\'${CONFIG_TYPE}\'","configName":"\'${CONFIG_NAME}\'","emailIds":"\'${EMAIL_IDS}\'"}''
    echo "------------FINISHING PLUGIN CUSTOM EMAIL NOTIFIER------------"
    ',
           'SHELL',
           'f',
           'now()',
           1,
           'now()',
           1
       );

-- Insert Plugin Step
INSERT INTO "plugin_step" ("id", "plugin_id", "name", "description", "index", "step_type", "script_id", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (
           nextval('id_seq_plugin_step'),
           (SELECT id FROM plugin_metadata WHERE name = 'Custom Email Notifier v1.0.0'),
           'Step 1',
           'Custom email notifier',
           1,
           'INLINE',
           (SELECT last_value FROM id_seq_plugin_pipeline_script),
           'f',
           'now()',
           1,
           'now()',
           1
       );

-- Insert Plugin Step Variables
INSERT INTO plugin_step_variable (id,plugin_step_id,name,format, description,is_exposed,allow_empty_value,default_value,value,variable_type,value_type,previous_step_index,variable_step_index,variable_step_index_in_plugin,reference_variable_name,deleted,created_on,created_by,updated_on,updated_by)VALUES
                                                                                                                                                                                                                                                                               (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Custom Email Notifier v1.0.0' and ps."index"=1 and ps.deleted=false),'CONFIG_TYPE','STRING','Type of the configuration','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
                                                                                                                                                                                                                                                                               (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Custom Email Notifier v1.0.0' and ps."index"=1 and ps.deleted=false),'CONFIG_NAME','STRING','Name of the configuration','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
                                                                                                                                                                                                                                                                               (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Custom Email Notifier v1.0.0' and ps."index"=1 and ps.deleted=false),'EMAIL_IDS','STRING','Email IDs to notify','t','t',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);