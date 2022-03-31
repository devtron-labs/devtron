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

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_tags;

-- Table Definition
CREATE TABLE "public"."plugin_tags"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_tags'::regclass),
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
    CONSTRAINT "plugin_tag_relation_tag_id_fkey" FOREIGN KEY ("tag_id") REFERENCES "public"."plugin_tags" ("id"),
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

CREATE SEQUENCE IF NOT EXISTS id_seq_script_path_arg_port_mappings;

-- Table Definition
CREATE TABLE "public"."script_path_arg_port_mappings"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_script_path_arg_port_mappings'::regclass),
    "type_of_mapping"             varchar(255),      -- FILE_PATH, DOCKER_ARG, PORT
    "file_path_on_disk"           text,
    "file_path_on_container"      text,
    "command"                     text,
    "arg"                         text,
    "port_on_local"               integer,
    "port_on_container"           integer,
    "script_id"                   integer,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "script_path_arg_port_mappings_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_steps;

-- Table Definition
CREATE TABLE "public"."plugin_steps"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_steps'::regclass),
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
    CONSTRAINT "plugin_steps_plugin_id_fkey" FOREIGN KEY ("plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    CONSTRAINT "plugin_steps_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    CONSTRAINT "plugin_steps_ref_plugin_id_fkey" FOREIGN KEY ("ref_plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step_variables;

-- Table Definition
CREATE TABLE "public"."plugin_step_variables"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_variables'::regclass),
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
    CONSTRAINT "plugin_step_variables_plugin_step_id_fkey" FOREIGN KEY ("plugin_step_id") REFERENCES "public"."plugin_steps" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_plugin_step_conditions;

-- Table Definition
CREATE TABLE "public"."plugin_step_conditions"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_step_conditions'::regclass),
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
    CONSTRAINT "plugin_step_conditions_plugin_step_id_fkey" FOREIGN KEY ("plugin_step_id") REFERENCES "public"."plugin_steps" ("id"),
    CONSTRAINT "plugin_step_conditions_condition_variable_id_fkey" FOREIGN KEY ("condition_variable_id") REFERENCES "public"."plugin_step_variables" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stages;

-- Table Definition
CREATE TABLE "public"."pipeline_stages"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stages'::regclass),
    "name"                        text,
    "description"                 text,
    "type"                        varchar(255),  -- PRE_CI, POST_CI, PRE_CD, POST_CD etc
    "icon"                        text,
    "deleted"                     bool,
    "ci_pipeline_id"              integer,
    "cd_pipeline_id"              integer,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stages_ci_pipeline_id_fkey" FOREIGN KEY ("ci_pipeline_id") REFERENCES "public"."ci_pipeline" ("id"),
    CONSTRAINT "pipeline_stages_cd_pipeline_id_fkey" FOREIGN KEY ("cd_pipeline_id") REFERENCES "public"."pipeline" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_steps;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_steps"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage_steps'::regclass),
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
    CONSTRAINT "pipeline_stage_steps_script_id_fkey" FOREIGN KEY ("script_id") REFERENCES "public"."plugin_pipeline_script" ("id"),
    CONSTRAINT "pipeline_stage_steps_ref_plugin_id_fkey" FOREIGN KEY ("ref_plugin_id") REFERENCES "public"."plugin_metadata" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step_variables;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step_variables"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_plugin_stage_step_variables'::regclass),
    "pipeline_stage_step_id"                   integer,
    "name"                        varchar(255),
    "format"                      varchar(255),
    "description"                 text,
    "is_exposed"                  bool,
    "allow_empty_value"           bool,
    "default_value"               varchar(255),
    "value"                       varchar(255),
    "variable_type"               varchar(255),   -- INPUT or OUTPUT
    "index"                       integer,
    "variable_value_type"         varchar(255),   -- NEW, FROM_PREVIOUS_STEP or GLOBAL
    "previous_step_index"         integer,
    "reference_variable_name"     text,
    "deleted"                     bool,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "pipeline_stage_step_variables_pipeline_stage_step_id_fkey" FOREIGN KEY ("pipeline_stage_step_id") REFERENCES "public"."pipeline_stage_steps" ("id"),
    PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_stage_step_conditions;

-- Table Definition
CREATE TABLE "public"."pipeline_stage_step_conditions"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_pipeline_stage_step_conditions'::regclass),
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
    CONSTRAINT "pipeline_stage_step_conditions_plugin_step_id_fkey" FOREIGN KEY ("pipeline_stage_step_id") REFERENCES "public"."pipeline_stage_steps" ("id"),
    CONSTRAINT "pipeline_stage_step_conditions_condition_variable_id_fkey" FOREIGN KEY ("condition_variable_id") REFERENCES "public"."pipeline_stage_step_variables" ("id"),
    PRIMARY KEY ("id")
);