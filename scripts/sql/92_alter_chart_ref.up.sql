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

UPDATE chart_ref set file_path_containing_strategy='pipeline-values.yaml' where user_uploaded=false;

UPDATE chart_ref set is_app_metrics_supported=true where version in ('3.7.0','3.8.0','3.9.0','3.10.0','3.11.0','3.12.0','3.13.0','4.10.0','4.11.0','4.12.0','4.13.0','4.14.0','4.15.0') and name is null and user_uploaded=false;

UPDATE chart_ref set is_app_metrics_supported=false where not (version in('3.7.0','3.8.0','3.9.0','3.10.0','3.11.0','3.12.0','3.13.0','4.10.0','4.11.0','4.12.0','4.13.0','4.14.0','4.15.0') and name is null and user_uploaded=false);

INSERT INTO global_strategy_metadata ("id","name", "description", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
    (1,'ROLLING', 'RollingUpdate or Rolling strategy.', 'false', 'now()', 1, 'now()', 1),
    (2,'BLUE-GREEN', 'Blue green strategy.', 'false', 'now()', 1, 'now()', 1),
    (3,'CANARY', 'Canary strategy.', 'false', 'now()', 1, 'now()', 1),
    (4,'RECREATE', 'Recreate strategy.', 'false', 'now()', 1, 'now()', 1);


SELECT pg_catalog.setval('public.id_seq_global_strategy_metadata', 4, true);

DO $$
DECLARE
charRefId varchar;
BEGIN
FOR chartRefId IN SELECT id FROM chart_ref where version in ('3.2.0','3.3.0','3.4.0','3.5.0','3.6.0','3.7.0','3.8.0','3.9.0','3.10.0','3.11.0','3.12.0','3.13.0','4.10.0','4.11.0','4.12.0','4.13.0','4.14.0','4.15.0') and name is null and user_uploaded=false
	LOOP
                  INSERT INTO id_seq_global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
                      (1,chartRefId, 'true', 'now()', 1, 'now()', 1),
                      (2,chartRefId,'true', 'now()', 1, 'now()', 1),
                      (3,chartRefId, 'true', 'now()', 1, 'now()', 1),
                      (4,chartRefId, 'true', 'now()', 1, 'now()', 1);
END LOOP;
END$$;


DO $$
DECLARE
charRefId varchar;
BEGIN
FOR chartRefId IN SELECT id FROM chart_ref where not (version in ('3.2.0','3.3.0','3.4.0','3.5.0','3.6.0','3.7.0','3.8.0','3.9.0','3.10.0','3.11.0','3.12.0','3.13.0','4.10.0','4.11.0','4.12.0','4.13.0','4.14.0','4.15.0') and name is null) and user_uploaded=false
    LOOP
                  INSERT INTO id_seq_global_strategy_metadata_chart_ref_mapping ("global_strategy_metadata_id", "chart_ref_id", "active", "created_on", "created_by", "updated_on", "updated_by") VALUES
                      (1,chartRefId, 'true', 'now()', 1, 'now()', 1),
                      (2,chartRefId,'true', 'now()', 1, 'now()', 1);
END LOOP;
END$$;