CREATE SEQUENCE IF NOT EXISTS id_seq_scan_tool_metadata;

CREATE TABLE public.scan_tool_metadata
(
    "id"                                  integer NOT NULL DEFAULT nextval('id_seq_scan_tool_metadata'::regclass),
    "name"                                VARCHAR(100),
    "version"                             VARCHAR(50),
    "server_base_url"                     text,
    "result_descriptor_template"            text,
    "scan_target"                          VARCHAR(10),
    "active"                              bool,
    "deleted"                               bool,
    "created_on"                          timestamptz,
    "created_by"                          int4,
    "updated_on"                          timestamptz,
    "updated_by"                          int4,
    "tool_metadata"                         text,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_scan_tool_execution_history_mapping;

CREATE TABLE public.scan_tool_execution_history_mapping
(
    "id"                             integer NOT NULL DEFAULT nextval('id_seq_scan_tool_execution_history_mapping'::regclass),
    "image_scan_execution_history_id" integer,
    "scan_tool_id"             integer,
    "execution_start_time"           timestamptz,
    "execution_finish_time"              timestamptz,
    "state"                            int,
    "try_count"                         int,
    "created_on"                     timestamptz,
    "created_by"                     int4,
    "updated_on"                     timestamptz,
    "updated_by"                     int4,
    PRIMARY KEY ("id"),
    CONSTRAINT "scan_tool_execution_history_mapping_result_id_fkey" FOREIGN KEY ("image_scan_execution_history_id") REFERENCES "public"."image_scan_execution_history" ("id"),
    CONSTRAINT "scan_tool_execution_history_mapping_scan_tool_id_fkey" FOREIGN KEY ("scan_tool_id") REFERENCES "public"."scan_tool_metadata" ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_scan_step_condition;

CREATE TABLE public.scan_step_condition
(
    "id"                   integer NOT NULL DEFAULT nextval('id_seq_scan_step_condition'::regclass),
    "condition_variable_format" VARCHAR(10),
    "conditional_operator" VARCHAR(5),
    "conditional_value"    VARCHAR(100),
    "condition_on"         text,
    "deleted"               bool,
    "created_on"           timestamptz,
    "created_by"           int4,
    "updated_on"           timestamptz,
    "updated_by"           int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_scan_tool_step;

CREATE TABLE public.scan_tool_step
(
    "id"                   integer NOT NULL DEFAULT nextval('id_seq_scan_tool_step'::regclass),
    "scan_tool_id"   integer,
    "index"                integer,
    "step_execution_type"  VARCHAR(10),
    "step_execution_sync"   bool NOT NULL ,
    "retry_count"          integer,
    "execute_step_on_fail" integer,
    "execute_step_on_pass" integer,
    "render_input_data_from_step" integer,
    "http_input_payload"        jsonb,
    "http_method_type"     text,
    "http_req_headers"     jsonb,
    "http_query_params"    jsonb,
    "cli_command"            text,
    "cli_output_type"      VARCHAR(10),
    "deleted"               bool,
    "created_on"           timestamptz,
    "created_by"           int4,
    "updated_on"           timestamptz,
    "updated_by"           int4,
    PRIMARY KEY ("id"),
    CONSTRAINT "scan_tool_step_scan_tool_id_fkey" FOREIGN KEY ("scan_tool_id") REFERENCES "public"."scan_tool_metadata" ("id")

);

CREATE SEQUENCE IF NOT EXISTS id_seq_scan_step_condition_mapping;

CREATE TABLE public.scan_step_condition_mapping
(
    "id"                           integer NOT NULL DEFAULT nextval('id_seq_scan_step_condition_mapping'::regclass),
    "scan_step_condition_id" integer,
    "scan_tool_step_id"      integer,
    "created_on"                   timestamptz,
    "created_by"                   int4,
    "updated_on"                   timestamptz,
    "updated_by"                   int4,
    PRIMARY KEY ("id"),
    CONSTRAINT "scan_step_condition_mapping_condition_id_fkey" FOREIGN KEY ("scan_step_condition_id") REFERENCES "public"."scan_step_condition" ("id"),
    CONSTRAINT "scan_step_condition_mapping_tool_step_id_fkey" FOREIGN KEY ("scan_tool_step_id") REFERENCES "public"."scan_tool_step" ("id")

);

CREATE SEQUENCE IF NOT EXISTS id_registry_index_mapping;

CREATE TABLE public.registry_index_mapping
(
    "id"                           integer NOT NULL DEFAULT nextval('id_registry_index_mapping'::regclass),
    "scan_tool_id"      integer,
    "registry_type"                   varchar(20),
    "starting_index"   integer,
    PRIMARY KEY ("id"),
    CONSTRAINT "registry_index_mapping_id_fkey" FOREIGN KEY ("scan_tool_id") REFERENCES "public"."scan_tool_metadata" ("id")
);

ALTER TABLE public.image_scan_execution_history
    ADD "scan_event_json" text ,
    ADD "execution_history_directory_path" text;

ALTER TABLE public.image_scan_execution_result
    ADD "scan_tool_id" int;