CREATE SEQUENCE IF NOT EXISTS id_seq_k8s_event_watcher;
CREATE TABLE "public"."k8s_event_watcher" (
                                    "id" integer NOT NULL default nextval('id_seq_k8s_event_watcher'::regclass),
                                    "name" varchar(50) NOT NULL,
                                    "description" text ,
                                    "filter_expression" text NOT NULL,
                                    "gvks" text,
                                    "selected_actions" varchar(15)[],
                                    "selectors" text,
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
                                              "search_data" text,
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
INSERT INTO "plugin_pipeline_script" ("id", "script", "type", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (
           nextval('id_seq_plugin_pipeline_script'),
           E'#!/bin/sh

    # URL and token from environment variables
    echo "------------STARTING PLUGIN CUSTOM EMAIL NOTIFIER------------"
    # Make the API call
   STRINGIFIED_JSON=$(echo "$NOTIFICATION_DATA" | jq -c .)
# Replace backslashes with escaped backslashes and double quotes with escaped double quotes
ESCAPED_JSON="${STRINGIFIED_JSON//\\\\/\\\\\\\\}"   # Replace backslashes
ESCAPED_JSON="${ESCAPED_JSON//\\"/\\\\\\"}"      # Replace double quotes
    curl -X POST ${NOTIFICATION_URL} -H "token: ${NOTIFICATION_TOKEN}" -H "Content-Type: application/json" -d ''{"configType": "\'${CONFIG_TYPE}\'","configName":"\'${CONFIG_NAME}\'","emailIds":"\'${EMAIL_IDS}\'","data":"\'${ESCAPED_JSON}\'"}''
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
    # Make the API call
    STRINGIFIED_JSON=$(echo "$NOTIFICATION_DATA" | jq -c .)
               # Replace backslashes with escaped backslashes and double quotes with escaped double quotes
               ESCAPED_JSON="${STRINGIFIED_JSON//\\\\/\\\\\\\\}"   # Replace backslashes
               ESCAPED_JSON="${ESCAPED_JSON//\\"/\\\\\\"}"      # Replace double quotes
    curl -X POST ${NOTIFICATION_URL} -H "token: $NOTIFICATION_TOKEN" -H "Content-Type: application/json" -d ''{"configType": "\'${CONFIG_TYPE}\'","configName":"\'${CONFIG_NAME}\'","data":"\'${ESCAPED_JSON}\'"}''
    echo "------------FINISHING PLUGIN CUSTOM WEBHOOK NOTIFIER------------"
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


-- insert notification templates
INSERT INTO public.event (id, event_type, description) VALUES (9, 'SCOOP RESOURCE INTERCEPT EVENT', '');
    INSERT INTO "public"."notification_templates" (channel_type, node_type, event_type_id, template_name, template_payload)
VALUES ('ses', 'CD', 9, 'scoop event ses template', '{
    "from": "{{fromEmail}}",
    "to": "{{toEmail}}",
    "subject": "🛎️ Change in resource intercepted | Resource {{action}} | Resource: {{kind}}/{{resourceName}}",
    "html":     "<table cellpadding=0 style=\"\n      \tfont-family: Arial, Verdana, Helvetica;\n\t\twidth: 600px;\n      height: 485px;\n      border-collapse: inherit;\n      border-spacing: 0;\n      border:1px solid #D0D4D9;\n\tborder-radius: 8px;\n      padding: 16px 20px;\n      margin: 20px auto;\n    \">\n    <tr>\n        <td colspan=\"3\">\n            <div style=\"height: 28px; padding-bottom: 16px; margin-bottom: 20px; border-bottom:1px solid #EDF1F5; max-width: 600px;\">\n                <img style=\"height: 100%\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\" />\n            </div>\n        </td>\n    </tr>\n    <tr>\n        <td colspan=\"3\">\n            <div style=\"background-color: {{color}}; border-top-left-radius: 8px; border-top-right-radius: 8px; padding: 20px 20px 16px 20px; display: flex; justify-content: space-between;\">\n                <div style=\"width:90%;\">\n                    <div style=\"font-size: 16px; line-height:24px; font-weight:600; margin-bottom:6px; color: #000a14;\">{{heading}}</div>\n                    <span style=\"font-size: 14px; line-height:20px; color: #000a14;\">{{kind}}/{{resourceName}}</span>\n                    <br/>\n                    <span style=\"font-size: 14px; line-height:20px; color: #000a14;\">{{interceptedAt}}</span>\n                </div>\n\t\t\t\t\n\t\t\t\t<div>\n                <img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height: 72px; width: 72px\"/>\n            </div>\n            </div>\n            \n        </td>\n\n    </tr>\n\t <tr>\n        <td colspan=\"3\">\n\t\t\t<div style=\"background-color: {{color}}; border-bottom-left-radius: 8px; border-bottom-right-radius: 8px; padding: 0 0 20px 20px;\">\n          <a\n              href=\"{{viewResourceManifestLink}}\"\n              style=\"\n                padding: 6px 12px;\n                line-height: 20px;\n                font-size: 12px;\n                font-weight: 600;\n                border-radius: 4px;\n                text-decoration: none;\n                outline: none;\n                min-width: 64px;\n                text-transform: capitalize;\n                text-align: center;\n                background: #0066cc;\n                color: #fff;\n                border: 1px solid transparent;\n                cursor: pointer;\n              \"\n              >View resource manifest</a\n            >\n\t\t\t\t</div>\n            \n        </td>\n\n    </tr>\n    <tr>\n        <tr>\n            <td>\n                <br>\n            </td>\n        </tr>\n        \n        <tr>\n            <td colspan=\"3\">\n                <div style=\"\n          font-weight: 600;\n          margin-top: 20px;\n          width: 100%;\n          border-top: 1px solid #EDF1F5;\n          padding: 16px 0;\n          font-size: 14px;\n      \">Details\n                </div>\n            </td>\n        </tr>\n        <tr>\n            <td>\n                <div style=\"color: #3B444C;\n            font-size: 13px; padding-bottom: 4px;\">Cluster</div>\n            </td>\n            <td>\n                <div style=\"color: #3B444C; font-size: 13px; padding-bottom: 4px;\">Namespace</div>\n            </td>\n        </tr>\n\n        <tr></tr>\n        <tr>\n            <td>\n                <div style=\"color: #000a14; font-size: 14px;\">{{clusterName}}</div>\n            </td>\n            <td>\n                <div style=\"color: #000a14; font-size: 14px;\">{{namespace}}</div>\n            </td>\n            <tr>\n                <td>\n                    <br>\n                </td>\n            </tr>\n        </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #3B444C; font-size: 13px; padding-bottom: 4px;\">Intercepted by watcher</div>\n            </td>\n        </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #000a14; font-size: 14px;\">{{watcherName}}</div</td>\n        </tr>\n    </tr>\n<tr>\n                <td>\n                    <br>\n                </td>\n            </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #3B444C; font-size: 13px; padding-bottom: 4px;\">Runbook (Job pipeline)</div>\n            </td>\n        </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #000a14; font-size: 14px;\">{{pipelineName}}</div</td>\n        </tr>\n    </tr>\n<tr>\n        <td colspan=\"3\">\n            <div style=\"border-top: 1px solid #EDF1F5; margin: 20px 0 16px 0; height: 1px\"></div>\n        </td>\n    </tr>\n    <tr>\n        <td colspan=\"2\" style=\"display: flex;\">\n            <span>\n              <a href=\"https://twitter.com/DevtronL\" target=\"_blank\"\n              style=\" cursor: pointer; text-decoration: none; padding-right: 12px; display: flex\">\n              <div>\n                <img style=\"width: 20px\" src=\"https://cdn.devtron.ai/images/twitter_social_dark.png\"/>\n                </div>\n            </a>\n          </span>\n            <span>\n          <a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\"\n          style=\" cursor: pointer; text-decoration: none; padding-right: 12px; display: flex\">\n          <div>\n            <img style=\"width: 20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"/>\n              </div>\n        </a>\n        </span>\n            <span>\n          <a href=\"https://devtron.ai/blog/\" target=\"_blank\"\n          style=\"color: #000a14; font-size:13px; line-height:20px; cursor: pointer; text-decoration: underline; padding-right: 12px;\">\n          Blog\n        </a>\n        </span>\n            <span>\n          <a href=\"https://devtron.ai/\" target=\"_blank\"\n          style=\"color: #000a14; font-size:13px; line-height:20px; cursor: pointer; text-decoration: underline;\">\n          Website\n        </a>\n         </span>\n        </td>\n        <td colspan=\"2\" style=\"text-align: right;\">\n            <div style=\"color: #767d84; font-size:13px; line-height:20px;\" >\n                &copy; Devtron Labs 2020\n            </div>\n        </td>\n    </tr>\n    </tr>\n</table>"
}');

INSERT INTO "public"."notification_templates" (channel_type, node_type, event_type_id, template_name, template_payload)
VALUES ('smtp', 'CD', 9, 'scoop event smtp template', '{
    "from": "{{fromEmail}}",
    "to": "{{toEmail}}",
    "subject": "🛎️ Change in resource intercepted | Resource {{action}} | Resource: {{kind}}/{{resourceName}}",
    "html":     "<table cellpadding=0 style=\"\n      \tfont-family: Arial, Verdana, Helvetica;\n\t\twidth: 600px;\n      height: 485px;\n      border-collapse: inherit;\n      border-spacing: 0;\n      border:1px solid #D0D4D9;\n\tborder-radius: 8px;\n      padding: 16px 20px;\n      margin: 20px auto;\n    \">\n    <tr>\n        <td colspan=\"3\">\n            <div style=\"height: 28px; padding-bottom: 16px; margin-bottom: 20px; border-bottom:1px solid #EDF1F5; max-width: 600px;\">\n                <img style=\"height: 100%\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\" />\n            </div>\n        </td>\n    </tr>\n    <tr>\n        <td colspan=\"3\">\n            <div style=\"background-color: {{color}}; border-top-left-radius: 8px; border-top-right-radius: 8px; padding: 20px 20px 16px 20px; display: flex; justify-content: space-between;\">\n                <div style=\"width:90%;\">\n                    <div style=\"font-size: 16px; line-height:24px; font-weight:600; margin-bottom:6px; color: #000a14;\">{{heading}}</div>\n                    <span style=\"font-size: 14px; line-height:20px; color: #000a14;\">{{kind}}/{{resourceName}}</span>\n                    <br/>\n                    <span style=\"font-size: 14px; line-height:20px; color: #000a14;\">{{interceptedAt}}</span>\n                </div>\n\t\t\t\t\n\t\t\t\t<div>\n                <img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height: 72px; width: 72px\"/>\n            </div>\n            </div>\n            \n        </td>\n\n    </tr>\n\t <tr>\n        <td colspan=\"3\">\n\t\t\t<div style=\"background-color: {{color}}; border-bottom-left-radius: 8px; border-bottom-right-radius: 8px; padding: 0 0 20px 20px;\">\n          <a\n              href=\"{{viewResourceManifestLink}}\"\n              style=\"\n                padding: 6px 12px;\n                line-height: 20px;\n                font-size: 12px;\n                font-weight: 600;\n                border-radius: 4px;\n                text-decoration: none;\n                outline: none;\n                min-width: 64px;\n                text-transform: capitalize;\n                text-align: center;\n                background: #0066cc;\n                color: #fff;\n                border: 1px solid transparent;\n                cursor: pointer;\n              \"\n              >View resource manifest</a\n            >\n\t\t\t\t</div>\n            \n        </td>\n\n    </tr>\n    <tr>\n        <tr>\n            <td>\n                <br>\n            </td>\n        </tr>\n        \n        <tr>\n            <td colspan=\"3\">\n                <div style=\"\n          font-weight: 600;\n          margin-top: 20px;\n          width: 100%;\n          border-top: 1px solid #EDF1F5;\n          padding: 16px 0;\n          font-size: 14px;\n      \">Details\n                </div>\n            </td>\n        </tr>\n        <tr>\n            <td>\n                <div style=\"color: #3B444C;\n            font-size: 13px; padding-bottom: 4px;\">Cluster</div>\n            </td>\n            <td>\n                <div style=\"color: #3B444C; font-size: 13px; padding-bottom: 4px;\">Namespace</div>\n            </td>\n        </tr>\n\n        <tr></tr>\n        <tr>\n            <td>\n                <div style=\"color: #000a14; font-size: 14px;\">{{clusterName}}</div>\n            </td>\n            <td>\n                <div style=\"color: #000a14; font-size: 14px;\">{{namespace}}</div>\n            </td>\n            <tr>\n                <td>\n                    <br>\n                </td>\n            </tr>\n        </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #3B444C; font-size: 13px; padding-bottom: 4px;\">Intercepted by watcher</div>\n            </td>\n        </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #000a14; font-size: 14px;\">{{watcherName}}</div</td>\n        </tr>\n    </tr>\n<tr>\n                <td>\n                    <br>\n                </td>\n            </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #3B444C; font-size: 13px; padding-bottom: 4px;\">Runbook (Job pipeline)</div>\n            </td>\n        </tr>\n        <tr>\n            <td colspan=\"3\">\n                <div style=\"color: #000a14; font-size: 14px;\">{{pipelineName}}</div</td>\n        </tr>\n    </tr>\n<tr>\n        <td colspan=\"3\">\n            <div style=\"border-top: 1px solid #EDF1F5; margin: 20px 0 16px 0; height: 1px\"></div>\n        </td>\n    </tr>\n    <tr>\n        <td colspan=\"2\" style=\"display: flex;\">\n            <span>\n              <a href=\"https://twitter.com/DevtronL\" target=\"_blank\"\n              style=\" cursor: pointer; text-decoration: none; padding-right: 12px; display: flex\">\n              <div>\n                <img style=\"width: 20px\" src=\"https://cdn.devtron.ai/images/twitter_social_dark.png\"/>\n                </div>\n            </a>\n          </span>\n            <span>\n          <a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\"\n          style=\" cursor: pointer; text-decoration: none; padding-right: 12px; display: flex\">\n          <div>\n            <img style=\"width: 20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"/>\n              </div>\n        </a>\n        </span>\n            <span>\n          <a href=\"https://devtron.ai/blog/\" target=\"_blank\"\n          style=\"color: #000a14; font-size:13px; line-height:20px; cursor: pointer; text-decoration: underline; padding-right: 12px;\">\n          Blog\n        </a>\n        </span>\n            <span>\n          <a href=\"https://devtron.ai/\" target=\"_blank\"\n          style=\"color: #000a14; font-size:13px; line-height:20px; cursor: pointer; text-decoration: underline;\">\n          Website\n        </a>\n         </span>\n        </td>\n        <td colspan=\"2\" style=\"text-align: right;\">\n            <div style=\"color: #767d84; font-size:13px; line-height:20px;\" >\n                &copy; Devtron Labs 2020\n            </div>\n        </td>\n    </tr>\n    </tr>\n</table>"
}');