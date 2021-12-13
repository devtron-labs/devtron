CREATE SEQUENCE IF NOT EXISTS id_seq_config_map_global_history;

-- Table Definition
CREATE TABLE "public"."config_map_global_history"
(
    "id"                          integer      NOT NULL DEFAULT nextval('id_seq_config_map_global_history'::regclass),
    "config_map_app_level_id"     integer      NOT NULL,
    "data_type"                   varchar(255),
    "data"                        text,
    "latest"                      boolean DEFAULT false NOT NULL,
    "deployed"                    boolean,
    "deployed_on"                 timestamptz,
    "deployed_by"                 int4,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "config_map_global_history_config_map_app_level_id_fkey" FOREIGN KEY ("config_map_app_level_id") REFERENCES "public"."config_map_app_level" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_config_map_env_history;

-- Table Definition
CREATE TABLE "public"."config_map_env_history"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_config_map_env_history'::regclass),
    "config_map_env_level_id"     integer NOT NULL,
    "data_type"                   varchar(255),
    "data"                        text,
    "latest"                      boolean DEFAULT false NOT NULL,
    "deployed"                    boolean,
    "deployed_on"                 timestamptz,
    "deployed_by"                 int4,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "config_map_env_history_config_map_env_level_id_fkey" FOREIGN KEY ("config_map_env_level_id") REFERENCES "public"."config_map_env_level" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_charts_global_history;

-- Table Definition
CREATE TABLE "public"."charts_global_history"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_charts_global_history'::regclass),
    "charts_id"                   integer NOT NULL,
    "values_yaml"                 text NOT NULL,
    "global_override"             text NOT NULL,
    "release_override"            text NOT NULL,
    "pipeline_override"           text DEFAULT '{}'::text NOT NULL,
    "image_descriptor_template"   text,
    "latest"                      boolean DEFAULT false NOT NULL,
    "chart_ref_id"                integer NOT NULL,
    "deployed"                    bool,
    "deployed_on"                 timestamptz,
    "deployed_by"                 int4,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "charts_global_history_charts_id_fkey" FOREIGN KEY ("charts_id") REFERENCES "public"."charts" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_charts_env_history;

-- Table Definition
CREATE TABLE "public"."charts_env_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_charts_env_history'::regclass),
    "chart_env_config_override_id"  integer NOT NULL,
    "target_environment"            integer,
    "env_override"                  text NOT NULL,
    "latest"                        boolean DEFAULT false NOT NULL,
    "deployed"                      bool,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "charts_env_history_chart_env_config_override_id_fkey" FOREIGN KEY ("chart_env_config_override_id") REFERENCES "public"."chart_env_config_override" ("id"),
    PRIMARY KEY ("id")
);