CREATE SEQUENCE IF NOT EXISTS id_seq_config_map_history;

-- Table Definition
CREATE TABLE "public"."config_map_history"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_config_map_history'::regclass),
    "pipeline_id"                 integer NOT NULL,
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

CREATE SEQUENCE IF NOT EXISTS id_seq_charts_history;

-- Table Definition
CREATE TABLE "public"."charts_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_charts_history'::regclass),
    "pipeline_id"                      integer NOT NULL,
    "target_environment"            integer,
    "image_descriptor_template"                      text NOT NULL,
    "template"                      text NOT NULL,
    "deployed"                      bool,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "charts_history_pipeline_id_fkey" FOREIGN KEY ("pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_installed_app_history;

-- Table Definition
CREATE TABLE "public"."installed_app_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_installed_app_history'::regclass),
    "installed_app_version_id"      integer NOT NULL,
    "values_yaml"                   text,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "installed_app_history_installed_app_version_id_fkey" FOREIGN KEY ("installed_app_version_id") REFERENCES "public"."installed_app_versions" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_ci_script_history;

-- Table Definition
CREATE TABLE "public"."ci_script_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_ci_script_history'::regclass),
    "ci_pipeline_scripts_id"        integer NOT NULL,
    "script"                        text,
    "stage"                         text,
    "built"                         bool,
    "built_on"                      timestamptz,
    "built_by"                      int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "ci_script_history_ci_pipeline_scripts_id_fkey" FOREIGN KEY ("ci_pipeline_scripts_id") REFERENCES "public"."ci_pipeline_scripts" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_cd_config_history;

-- Table Definition
CREATE TABLE "public"."cd_config_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_cd_config_history'::regclass),
    "pipeline_id"                   integer NOT NULL,
    "config"                        text,
    "stage"                         text,
    "configmap_secret_names"        text,
    "deployed"                      bool,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "cd_config_history_pipeline_id_fkey" FOREIGN KEY ("pipeline_id") REFERENCES "public"."pipeline" ("id"),
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