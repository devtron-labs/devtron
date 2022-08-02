CREATE SEQUENCE IF NOT EXISTS id_seq_cd_pipeline_status_timeline;

-- Table Definition
CREATE TABLE "public"."cd_pipeline_status_timeline"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_cd_pipeline_status_timeline'::regclass),
    "status"                      varchar(255),
    "status_detail"               text,
    "status_time"                 timestamptz,
    "cd_workflow_runner_id"       integer,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "cd_pipeline_status_timeline_cd_workflow_runner_id_fkey" FOREIGN KEY ("cd_workflow_runner_id") REFERENCES "public"."cd_workflow_runner" ("id"),
    PRIMARY KEY ("id")
);