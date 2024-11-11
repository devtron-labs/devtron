CREATE SEQUENCE IF NOT EXISTS id_seq_deployment_event;
CREATE TABLE IF NOT EXISTS public.deployment_event
(
    "id"                           int          NOT NULL DEFAULT nextval('id_seq_deployment_event'::regclass),
    "app_id"                       int,
    "env_id"                       int,
    "pipeline_id"                  int,
    "cd_workflow_runner_id"        int,
    "event_json"                   text          NOT NULL,
    "metadata"                   text          NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id")
    );


Alter table devtron_resource_task_run alter column run_source_dependency_identifier drop not null;
Alter table devtron_resource_task_run alter column task_type_identifier drop not null;