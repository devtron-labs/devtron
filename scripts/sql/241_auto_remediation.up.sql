CREATE SEQUENCE IF NOT EXISTS id_seq_k8s_event_watcher;
CREATE TABLE "public"."k8s_event_watcher" (
                                    "id" integer NOT NULL default nextval('id_seq_k8s_event_watcher'::regclass),
                                    "name" varchar(50) NOT NULL,
                                    "description" text ,
                                    "filter_expression" text NOT NULL,
                                    "gvks" text,
                                    "selected_actions" []varchar(15),
                                    "active" bool NOT NULL,
                                    "created_on"                timestamptz NOT NULL,
                                    "created_by"                int4        NOT NULL,
                                    "updated_on"                timestamptz,
                                    "updated_by"                int4,

                                    PRIMARY KEY ("id")

);

CREATE UNIQUE INDEX "idx_unique_k8s_event_watcher_name"
    ON k8s_event_watcher(name)
    WHERE  k8s_event_watcher.active =true;

CREATE SEQUENCE IF NOT EXISTS id_seq_auto_remediation_trigger;
CREATE TABLE "public"."auto_remediation_trigger"(
                                   "id" integer NOT NULL default nextval('id_seq_auto_remediation_trigger'::regclass),
                                   "type" varchar(50) , -- DEVTRON_JOB
                                   "watcher_id" integer ,
                                   "data" text,
                                   "active" bool NOT NULL,
                                   "created_on"                timestamptz NOT NULL,
                                   "created_by"                int4        NOT NULL,
                                   "updated_on"                timestamptz,
                                   "updated_by"                int4,

                                   CONSTRAINT auto_remediation_trigger_k8s_event_watcher_id_fkey FOREIGN KEY ("watcher_id") REFERENCES "public"."k8s_event_watcher" ("id"),
                                   PRIMARY KEY ("id")

);

CREATE SEQUENCE IF NOT EXISTS id_seq_intercepted_event_execution;
CREATE TABLE "public"."intercepted_event_execution"(
                                              "id" integer NOT NULL default nextval('id_seq_intercepted_event_execution'::regclass),
                                              "cluster_id" int ,
                                              "namespace" character varying(250) NOT NULL,
                                              "metadata" text,
                                              "execution_message" text,
                                              "action" varchar(20),
                                              "involved_objects" text,
                                              "intercepted_at" timestamptz,
                                              "status" varchar(32) , -- failure,success,inprogress
                                              "trigger_id" integer,
                                              "trigger_execution_id" integer,
                                              "created_on"                timestamptz NOT NULL,
                                              "created_by"                int4        NOT NULL,
                                              "updated_on"                timestamptz,
                                              "updated_by"                int4,

                                              CONSTRAINT intercepted_events_auto_remediation_trigger_id_fkey FOREIGN KEY ("trigger_id") REFERENCES "public"."auto_remediation_trigger" ("id"),
                                              CONSTRAINT intercepted_events_cluster_id_fkey FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster" ("id"),
                                              PRIMARY KEY ("id")
);


-- PLUGIN SCRIPT

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
    curl -X POST ${NOTIFICATION_URL} -H "token: ${NOTIFICATION_TOKEN}" -H "Content-Type: application/json" -d ''{"configType": "\'${CONFIG_TYPE}\'","configName":"\'${CONFIG_NAME}\'","emailIds":"\'${EMAIL_IDS}\'"}''
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


-- Insert Plugin Metadata
INSERT INTO "plugin_metadata" ("id", "name", "description", "type", "icon", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (nextval('id_seq_plugin_metadata'), 'Custom Webhook Notifier v1.0.0', 'Send webhook notifications', 'PRESET', 'https://seeklogo.com/images/W/webhooks-logo-04229CC4AE-seeklogo.com.png', 'f', 'now()', 1, 'now()', 1);

-- Insert Plugin Stage Mapping
INSERT INTO "plugin_stage_mapping" ("plugin_id", "stage_type", "created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name = 'Custom Webhook Notifier v1.0.0'), 0, 'now()', 1, 'now()', 1);

-- Insert Plugin Script
INSERT INTO "plugin_pipeline_script" ("id", "script", "type", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (
           nextval('id_seq_plugin_pipeline_script'),
           E'#!/bin/sh

    # URL and token from environment variables
    echo "------------STARTING PLUGIN CUSTOM WEBHOOK NOTIFIER------------"
    echo "CONFIG_TYPE: $CONFIG_TYPE"
    echo "CONFIG_NAME: $CONFIG_NAME"
    url = $NOTIFICATION_URL
    token = $NOTIFICATION_TOKEN
    echo "NOTIFICATION_URL: $NOTIFICATION_URL"
    echo "NOTIFICATION_TOKEN: $NOTIFICATION_TOKEN"
    echo "{"configType": \'"${CONFIG_TYPE}"\', "configName": \'"${CONFIG_NAME}"\'}"
    echo \'{"configType":\'${CONFIG_TYPE}\',"configName":\'${CONFIG_NAME}\'}\'
    # Make the API call
    curl -X POST ${NOTIFICATION_URL} -H "token: ${NOTIFICATION_TOKEN}" -H "Content-Type: application/json" -d ''{"configType": "\'${CONFIG_TYPE}\'","configName":"\'${CONFIG_NAME}\'"}''
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
           (SELECT id FROM plugin_metadata WHERE name = 'Custom Webhook Notifier v1.0.0'),
           'Step 1',
           'Custom webhook notifier',
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
                                                                                                                                                                                                                                                                                                             (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Custom Webhook Notifier v1.0.0' and ps."index"=1 and ps.deleted=false),'CONFIG_TYPE','STRING','Type of the configuration','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1),
                                                                                                                                                                                                                                                                                                             (nextval('id_seq_plugin_step_variable'),(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Custom Webhook Notifier v1.0.0' and ps."index"=1 and ps.deleted=false),'CONFIG_NAME','STRING','Name of the configuration','t','f',null,null,'INPUT','NEW',null,1,null,null,'f','now()',1,'now()',1);