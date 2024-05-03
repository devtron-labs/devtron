CREATE SEQUENCE IF NOT EXISTS id_seq_k8s_event_watcher;
CREATE TABLE "public"."k8s_event_watcher" (
                                    "id" integer NOT NULL default nextval('id_seq_k8s_event_watcher'::regclass),
                                    "name" varchar(50) NOT NULL,
                                    "description" text ,
                                    "filter_expression" text NOT NULL,
                                    "gvks" text,
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
                                              "gvk" text,
                                              "execution_message" text,
                                              "action" varchar(20),
                                              "involved_object" text,
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
VALUES (nextval('id_seq_plugin_metadata'), 'Custom Notification v1.0.0', 'Send notifications', 'PRESET', 'https://static.vecteezy.com/system/resources/previews/009/394/760/non_2x/bell-icon-transparent-notification-free-png.png', 'f', 'now()', 1, 'now()', 1);

-- Insert Plugin Stage Mapping
INSERT INTO "plugin_stage_mapping" ("plugin_id", "stage_type", "created_on", "created_by", "updated_on", "updated_by")
VALUES ((SELECT id FROM plugin_metadata WHERE name = 'Custom Notification v1.0.0'), 0, 'now()', 1, 'now()', 1);

-- Insert Plugin Script
INSERT INTO "plugin_pipeline_script" ("id", "script", "type", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES (
           nextval('id_seq_plugin_pipeline_script'),
           E'#!/bin/bash
    set -eo pipefail

    # URL and token from environment variables
    url="NOTIFICATION_URL"
    token="NOTIFICATION_TOKEN"

    # Make the API call
    curl --location "$url" \
    --header "token: $token" \
    --header "Content-Type: application/json" \
    --data-raw '{
           "configType": "$CONFIG_TYPE",
           "configName": "$CONFIG_NAME",
           "emailIds": "$EMAIL_IDS"
    }'
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
           (SELECT id FROM plugin_metadata WHERE name = 'Custom Notification v1.0.0'),
           'Step 1',
           'Custom notifier',
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
INSERT INTO "plugin_step_variable" ("id", "plugin_step_id", "name", "format", "description", "is_exposed", "allow_empty_value", "variable_type", "value_type", "variable_step_index", "deleted", "created_on", "created_by", "updated_on", "updated_by")
VALUES
    (
        nextval('id_seq_plugin_step_variable'),
        (SELECT ps.id FROM plugin_metadata p INNER JOIN plugin_step ps ON ps.plugin_id = p.id WHERE p.name = 'Custom Notification v1.0.0' AND ps.index = 1 AND ps.deleted = false),
        'CONFIG_TYPE',
        'STRING',
        'Type of the configuration',
        true,
        false,
        'INPUT',
        'NEW',
        1,
        'f',
        'now()',
        1,
        'now()',
        1
    ),
    (
        nextval('id_seq_plugin_step_variable'),
        (SELECT ps.id FROM plugin_metadata p INNER JOIN plugin_step ps ON ps.plugin_id = p.id WHERE p.name = 'Custom Notification v1.0.0' AND ps.index = 1 AND ps.deleted = false),
        'CONFIG_NAME',
        'STRING',
        'Name of the configuration',
        true,
        false,
        'INPUT',
        'NEW',
        1,
        'f',
        'now()',
        1,
        'now()',
        1
    ),
    (
        nextval('id_seq_plugin_step_variable'),
        (SELECT ps.id FROM plugin_metadata p INNER JOIN plugin_step ps ON ps.plugin_id = p.id WHERE p.name = 'Custom Notification v1.0.0' AND ps.index = 1 AND ps.deleted = false),
        'EMAIL_IDS',
        'STRING',
        'Email IDs to notify',
        true,
        true,
        'INPUT',
        'NEW',
        1,
        'f',
        'now()',
        1,
        'now()',
        1
    );
