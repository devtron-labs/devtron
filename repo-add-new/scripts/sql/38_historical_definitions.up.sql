CREATE SEQUENCE IF NOT EXISTS id_seq_config_map_history;

-- Table Definition
CREATE TABLE "public"."config_map_history"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_config_map_history'::regclass),
    "pipeline_id"                 integer,
    "app_id"                      integer,
    "data_type"                   varchar(255),
    "data"                        text,
    "deployed"                    boolean,
    "deployed_on"                 timestamptz,
    "deployed_by"                 int4,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "config_map_history_pipeline_id_fkey" FOREIGN KEY ("pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_deployment_template_history;

-- Table Definition
CREATE TABLE "public"."deployment_template_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_deployment_template_history'::regclass),
    "pipeline_id"                   integer,
    "app_id"                        integer,
    "target_environment"            integer,
    "image_descriptor_template"     text NOT NULL,
    "template"                      text NOT NULL,
    "template_name"                 text,
    "template_version"              text,
    "is_app_metrics_enabled"        bool,
    "deployed"                      bool,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "deployment_template_history_pipeline_id_fkey" FOREIGN KEY ("pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_app_store_charts_history;

-- Table Definition
CREATE TABLE "public"."app_store_charts_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_app_store_charts_history'::regclass),
    "installed_apps_id"             integer NOT NULL,
    "values_yaml"                   text,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "app_store_charts_history_installed_apps_id_fkey" FOREIGN KEY ("installed_apps_id") REFERENCES "public"."installed_apps" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pre_post_ci_script_history;

-- Table Definition
CREATE TABLE "public"."pre_post_ci_script_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_pre_post_ci_script_history'::regclass),
    "ci_pipeline_scripts_id"        integer NOT NULL,
    "script"                        text,
    "stage"                         text,
    "name"                          text,
    "output_location"               text,
    "built"                         bool,
    "built_on"                      timestamptz,
    "built_by"                      int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "pre_post_ci_script_history_ci_pipeline_scripts_id_fkey" FOREIGN KEY ("ci_pipeline_scripts_id") REFERENCES "public"."ci_pipeline_scripts" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pre_post_cd_script_history;

-- Table Definition
CREATE TABLE "public"."pre_post_cd_script_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_pre_post_cd_script_history'::regclass),
    "pipeline_id"                   integer NOT NULL,
    "script"                        text,
    "stage"                         text,
    "configmap_secret_names"        text,
    "configmap_data"                text,
    "secret_data"                   text,
    "exec_in_env"                   bool,
    "trigger_type"                  text,
    "deployed"                      bool,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "pre_post_cd_script_history_pipeline_id_fkey" FOREIGN KEY ("pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_strategy_history;

-- Table Definition
CREATE TABLE "public"."pipeline_strategy_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_pipeline_strategy_history'::regclass),
    "pipeline_id"                   integer NOT NULL,
    "config"                        text,
    "strategy"                      text NOT NULL ,
    "default"                       bool,
    "deployed"                      bool,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "pipeline_strategy_history_pipeline_id_fkey" FOREIGN KEY ("pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);