CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_metadata;

-- Table Definition
CREATE TABLE "public"."plugin_metadata"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_metadata'::regclass),
    "name"                        text,
    "description"                 text,
    "type"                        varchar(255),  -- SHARED, PRESET etc
    "icon"                        text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_tag;

-- Table Definition
CREATE TABLE "public"."plugin_tag"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_tag'::regclass),
    "name"                        varchar(255),
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_tag_relation;

-- Table Definition
CREATE TABLE "public"."plugin_tag_relation"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_tag_relation'::regclass),
    "tag_id"                      integer,
    "plugin_id"                   integer,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "plugin_tag_relation_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "public"."plugin_tag" ("id"),
    CONSTRAINT "plugin_tag_relation_plugin_id_fkey" FOREIGN KEY ("plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_pipeline_script;

-- Table Definition
CREATE TABLE "public"."plugin_pipeline_script"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_pipeline_script'::regclass),
    "script"                      text,
    "type"                        varchar(255),   -- SHELL, DOCKERFILE, CONTAINER_IMAGE etc
    "dockerfile_exists"           bool,
    "store_script_at"             text,
    "mount_path"                  text,
    "mount_code_to_container"     bool,
    "configure_mount_path"        bool,
    "container_image_path"        text,
    "image_pull_secret_type"      varchar(255),   -- CONTAINER_REGISTRY or SECRET_PATH
    "image_pull_secret"           text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_script_path_arg_port_mapping;

-- Table Definition
CREATE TABLE "public"."script_path_arg_port_mapping"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_script_path_arg_port_mapping'::regclass),
    "type_of_mapping"             varchar(255),      -- FILE_PATH, DOCKER_ARG, PORT
    "file_path_on_disk"           text,
    "file_path_on_container"      text,
    "command"                     text,
    "arg"                         text,
    "port_on_local"               integer,
    "port_on_container"           integer,
    "script_id"                   integer,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "script_path_arg_port_mapping_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step;

-- Table Definition
CREATE TABLE "public"."plugin_step"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_step'::regclass),
    "plugin_id"                   integer,        -- id of plugin - parent of this step
    "name"                        varchar(255),
    "description"                 text,
    "index"                       integer,
    "step_type"                   varchar(255),   -- INLINE or REF_PLUGIN
    "script_id"                   integer,
    "ref_plugin_id"               integer,        -- id of plugin used as reference
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "plugin_step_plugin_id_fkey" FOREIGN KEY ("plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    CONSTRAINT "plugin_step_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    CONSTRAINT "plugin_step_ref_plugin_id_fkey" FOREIGN KEY ("ref_plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step_variable;

-- Table Definition
CREATE TABLE "public"."plugin_step_variable"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_variable'::regclass),
    "plugin_step_id"              integer,
    "name"                        varchar(255),
    "format"                      varchar(255),
    "description"                 text,
    "is_exposed"                  bool,
    "allow_empty_value"           bool,
    "default_value"               varchar(255),
    "value"                       varchar(255),
    "variable_type"               varchar(255),   -- INPUT or OUTPUT
    "value_type"                  varchar(255),   -- NEW, FROM_PREVIOUS_STEP or GLOBAL
    "previous_step_index"         integer,
    "reference_variable_name"     text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "plugin_step_variable_plugin_step_id_fkey" FOREIGN KEY ("plugin_step_id") REFERENCES "public"."plugin_step" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step_condition;

-- Table Definition
CREATE TABLE "public"."plugin_step_condition"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_step_condition'::regclass),
    "plugin_step_id"              integer,
    "condition_variable_id"       integer,      -- id of variable on which condition is written
    "condition_type"              varchar(255), -- SKIP, TRIGGER, SUCCESS or FAILURE
    "conditional_operator"        varchar(255),
    "conditional_value"           varchar(255),
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "plugin_step_condition_plugin_step_id_fkey" FOREIGN KEY ("plugin_step_id") REFERENCES "public"."plugin_step" ("id"),
    CONSTRAINT "plugin_step_condition_condition_variable_id_fkey" FOREIGN KEY ("condition_variable_id") REFERENCES "public"."plugin_step_variable" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage;

-- Table Definition
CREATE TABLE "public"."pipeline_stage"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage'::regclass),
    "name"                        text,
    "description"                 text,
    "type"                        varchar(255),  -- PRE_CI, POST_CI, PRE_CD, POST_CD etc
    "deleted"                     bool,
    "ci_pipeline_id"              integer,
    "cd_pipeline_id"              integer,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_ci_pipeline_id_fkey" FOREIGN KEY ("ci_pipeline_id") REFERENCES "public"."ci_pipeline" ("id"),
    CONSTRAINT "pipeline_stage_cd_pipeline_id_fkey" FOREIGN KEY ("cd_pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage_step'::regclass),
    "pipeline_stage_id"           integer,
    "name"                        varchar(255),
    "description"                 text,
    "index"                       integer,
    "step_type"                   varchar(255),   -- INLINE or REF_PLUGIN
    "script_id"                   integer,
    "ref_plugin_id"               integer,        -- id of plugin used as reference
    "report_directory_path"       text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_step_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    CONSTRAINT "pipeline_stage_step_ref_plugin_id_fkey" FOREIGN KEY ("ref_plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step_variable;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step_variable"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_stage_step_variable'::regclass),
    "pipeline_stage_step_id"      integer,
    "name"                        varchar(255),
    "format"                      varchar(255),
    "description"                 text,
    "is_exposed"                  bool,
    "allow_empty_value"           bool,
    "default_value"               varchar(255),
    "value"                       varchar(255),
    "variable_type"               varchar(255),   -- INPUT or OUTPUT
    "index"                       integer,
    "value_type"                  varchar(255),   -- NEW, FROM_PREVIOUS_STEP or GLOBAL
    "previous_step_index"         integer,
    "reference_variable_name"     text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_step_variable_pipeline_stage_step_id_fkey" FOREIGN KEY ("pipeline_stage_step_id") REFERENCES "public"."pipeline_stage_step" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step_condition;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step_condition"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage_step_condition'::regclass),
    "pipeline_stage_step_id"      integer,
    "condition_variable_id"       integer,      -- id of variable on which condition is written
    "condition_type"              varchar(255), -- SKIP, TRIGGER, SUCCESS or FAILURE
    "conditional_operator"        varchar(255),
    "conditional_value"           varchar(255),
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_step_condition_plugin_step_id_fkey" FOREIGN KEY ("pipeline_stage_step_id") REFERENCES "public"."pipeline_stage_step" ("id"),
    CONSTRAINT "pipeline_stage_step_condition_condition_variable_id_fkey" FOREIGN KEY ("condition_variable_id") REFERENCES "public"."pipeline_stage_step_variable" ("id"),
    PRIMARY KEY ("id")
);