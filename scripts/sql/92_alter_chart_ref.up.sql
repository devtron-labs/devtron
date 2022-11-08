ALTER TABLE "chart_ref" ADD COLUMN "file_path_containing_strategy" text;
ALTER TABLE "chart_ref" ADD COLUMN "json_path_for_strategy" text;
ALTER TABLE "chart_ref" ADD COLUMN "is_app_metrics_supported" bool NOT NULL DEFAULT TRUE;

CREATE SEQUENCE IF NOT EXISTS id_seq_global_strategy_metadata;

CREATE TABLE public.global_strategy_metadata (
"id"                            integer NOT NULL DEFAULT nextval('id_seq_global_strategy_metadata'::regclass),
"name"                          text,
"description"                   text,
"deleted"                       bool NOT NULL DEFAULT FALSE,
"created_on"                    timestamptz,
"created_by"                    int4,
"updated_on"                    timestamptz,
"updated_by"                    int4,
PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_global_strategy_metadata_chart_ref_mapping;

CREATE TABLE public.global_strategy_metadata_chart_ref_mapping (
"id"                            integer NOT NULL DEFAULT nextval('id_seq_global_strategy_metadata_chart_ref_mapping'::regclass),
"global_strategy_metadata_id"   integer,
"chart_ref_id"                  integer,
"active"                        bool NOT NULL DEFAULT TRUE,
"created_on"                    timestamptz,
"created_by"                    int4,
"updated_on"                    timestamptz,
"updated_by"                    int4,
PRIMARY KEY ("id")
);


UPDATE chart_ref set is_app_metrics_supported=true where version in ('3.7.0','3.8.0','3.9.0','3.10.0','3.11.0','3.12.0','3.13.0','4.10.0','4.11.0','4.12.0','4.13.0','4.14.0','4.15.0') and name is null;

UPDATE chart_ref set is_app_metrics_supported=false where not (version in('3.7.0','3.8.0','3.9.0','3.10.0','3.11.0','3.12.0','3.13.0','4.10.0','4.11.0','4.12.0','4.13.0','4.14.0','4.15.0') and name is null);
