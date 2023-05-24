CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_status_timeline_resources;

CREATE TABLE public.pipeline_status_timeline_resources (
"id"                                integer NOT NULL DEFAULT nextval('id_seq_pipeline_status_timeline_resources'::regclass),
"installed_app_version_history_id"  integer,
"cd_workflow_runner_id"             integer,
"resource_name"                     VARCHAR(1000),
"resource_kind"                     VARCHAR(1000),
"resource_group"                    VARCHAR(1000),
"resource_phase"                    text,
"resource_status"                   text,
"status_message"                    text,
"timeline_stage"                    VARCHAR(100) DEFAULT 'KUBECTL_APPLY',
"created_on"                        timestamptz,
"created_by"                        int4,
"updated_on"                        timestamptz,
"updated_by"                        int4,
CONSTRAINT "pipeline_status_timeline_resources_cd_workflow_runner_id_fkey" FOREIGN KEY ("cd_workflow_runner_id") REFERENCES "public"."cd_workflow_runner" ("id"),
CONSTRAINT "pipeline_status_timeline_resources_installed_app_version_history_id_fkey" FOREIGN KEY ("installed_app_version_history_id") REFERENCES "public"."installed_app_version_history" ("id"),
PRIMARY KEY ("id")
);


CREATE SEQUENCE IF NOT EXISTS id_seq_pipeline_status_timeline_sync_detail;

CREATE TABLE public.pipeline_status_timeline_sync_detail (
"id"                                integer NOT NULL DEFAULT nextval('id_seq_pipeline_status_timeline_sync_detail'::regclass),
"installed_app_version_history_id"  integer,
"cd_workflow_runner_id"             integer,
"last_synced_at"                    timestamptz,
"sync_count"                        integer,
"created_on"                        timestamptz,
"created_by"                        int4,
"updated_on"                        timestamptz,
"updated_by"                        int4,
 CONSTRAINT "pipeline_status_timeline_sync_detail_cd_workflow_runner_id_fkey" FOREIGN KEY ("cd_workflow_runner_id") REFERENCES "public"."cd_workflow_runner" ("id"),
 CONSTRAINT "pipeline_status_timeline_sync_detail_installed_app_version_history_id_fkey" FOREIGN KEY ("installed_app_version_history_id") REFERENCES "public"."installed_app_version_history" ("id"),
 PRIMARY KEY ("id")
);

ALTER TABLE pipeline ADD COLUMN deployment_app_name text;

DO $$
DECLARE
temprow record;
BEGIN
    FOR temprow IN SELECT p.id, a.app_name, e.environment_name FROM pipeline p INNER JOIN app a on p.app_id = a.id INNER JOIN environment e on p.environment_id = e.id and p.deleted=false
        LOOP
            UPDATE pipeline SET deployment_app_name=FORMAT('%s-%s',temprow.app_name,temprow.environment_name) where id=temprow.id;
        END LOOP;
END$$;

ALTER TABLE cd_workflow_runner ADD COLUMN created_on timestamptz;

ALTER TABLE cd_workflow_runner ADD COLUMN created_by int4;

ALTER TABLE cd_workflow_runner ADD COLUMN updated_on timestamptz;

ALTER TABLE cd_workflow_runner ADD COLUMN updated_by int4;

DO $$
DECLARE
temprow record;
BEGIN
FOR temprow IN SELECT * FROM cd_workflow_runner
    LOOP
UPDATE cd_workflow_runner SET created_on=temprow.started_on, created_by=1, updated_on=temprow.started_on, updated_by=1 where id=temprow.id;
END LOOP;
END$$;