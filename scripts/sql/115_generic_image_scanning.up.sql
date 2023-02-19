CREATE SEQUENCE IF NOT EXISTS id_seq_image_scan_tool_metadata;

CREATE TABLE public.image_scan_tool_metadata
(
    "id"                                  integer NOT NULL DEFAULT nextval('id_seq_image_scan_tool_metadata'::regclass),
    "name"                                text,
    "version"                             VARCHAR(50),
    "server_base_url"                     text,
    "result_descriptor_template_location" text,
    "active"                              bool,
    "created_on"                          timestamptz,
    "created_by"                          int4,
    "updated_on"                          timestamptz,
    "updated_by"                          int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_scan_execution_result_mapping;

CREATE TABLE public.scan_execution_result_mapping
(
    "id"                             integer NOT NULL DEFAULT nextval('id_seq_scan_execution_result_mapping'::regclass),
    "image_scan_execution_result_id" integer,
    "image_scan_tool_id"             integer,
    "created_on"                     timestamptz,
    "created_by"                     int4,
    "updated_on"                     timestamptz,
    "updated_by"                     int4,
    PRIMARY KEY ("id"),
    CONSTRAINT "scan_execution_result_mapping_result_id_fkey" FOREIGN KEY ("image_scan_execution_result_id") REFERENCES "public"."image_scan_execution_result" ("id"),
    CONSTRAINT "scan_execution_result_mapping_scan_tool_id_fkey" FOREIGN KEY ("image_scan_tool_id") REFERENCES "public"."image_scan_tool_metadata" ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_image_scan_step_condition;

CREATE TABLE public.image_scan_step_condition
(
    "id"                   integer NOT NULL DEFAULT nextval('id_seq_image_scan_step_condition'::regclass),
    "conditional_operator" VARCHAR(5),
    "conditional_value"    VARCHAR(100),
    "condition_on"         text,
    "active"               bool,
    "created_on"           timestamptz,
    "created_by"           int4,
    "updated_on"           timestamptz,
    "updated_by"           int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_image_scan_tool_step;

CREATE TABLE public.image_scan_tool_step
(
    "id"                   integer NOT NULL DEFAULT nextval('id_seq_image_scan_tool_step'::regclass),
    "image_scan_tool_id"   integer,
    "index"                integer,
    "step_execution_type"  VARCHAR(10),
    "retry_count"          integer,
    "execute_step_on_fail" integer,
    "execute_step_on_pass" integer,
    "input_payload"        jsonb,
    "http_req_url"         text,
    "http_req_type"        VARCHAR(6),
    "http_req_headers"     jsonb,
    "http_query_params"    jsonb,
    "cli_flags"            jsonb,
    "cli_output_type"      VARCHAR(10),
    "active"               bool,
    "created_on"           timestamptz,
    "created_by"           int4,
    "updated_on"           timestamptz,
    "updated_by"           int4,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_image_scan_step_condition_mapping;

CREATE TABLE public.image_scan_step_condition_mapping
(
    "id"                           integer NOT NULL DEFAULT nextval('id_seq_image_scan_step_condition_mapping'::regclass),
    "image_scan_step_condition_id" integer,
    "image_scan_tool_step_id"      integer,
    "created_on"                   timestamptz,
    "created_by"                   int4,
    "updated_on"                   timestamptz,
    "updated_by"                   int4,
    PRIMARY KEY ("id"),
    CONSTRAINT "image_scan_step_condition_mapping_condition_id_fkey" FOREIGN KEY ("image_scan_step_condition_id") REFERENCES "public"."image_scan_step_condition" ("id"),
    CONSTRAINT "image_scan_step_condition_mapping_tool_step_id_fkey" FOREIGN KEY ("image_scan_tool_step_id") REFERENCES "public"."image_scan_tool_step" ("id")

);
