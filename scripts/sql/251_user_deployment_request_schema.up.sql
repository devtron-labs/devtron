-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_user_deployment_request_sequence;


-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."user_deployment_request"
(
    "id"                                     integer NOT NULL DEFAULT nextval('id_seq_user_deployment_request_sequence'::regclass),
    "pipeline_id"                            integer NOT NULL,
    "ciArtifact_id"                          integer NOT NULL,
    "additional_override"                    bytea,
    "force_trigger"                          bool    NOT NULL DEFAULT FALSE,
    "force_sync"                             bool    NOT NULL DEFAULT FALSE,
    "strategy"                               varchar(100),
    "deployment_with_config"                 varchar(100),
    "specific_trigger_wfr_id"                integer,
    "cd_workflow_id"                         integer NOT NULL,
    "deployment_type"                        integer,
    "triggered_at"                           timestamptz NOT NULL,
    "triggered_by"                           int4 NOT NULL,
    "status"                                 varchar(100) NOT NULL,

    CONSTRAINT user_deployment_request_pipeline_id_fk
        FOREIGN KEY (pipeline_id)
            REFERENCES public.pipeline(id),
    CONSTRAINT user_deployment_request_ciArtifact_id_fk
        FOREIGN KEY (ciArtifact_id)
            REFERENCES public.ciArtifact(id),
    CONSTRAINT user_deployment_request_pipeline_id_fk
        FOREIGN KEY (cd_workflow_id)
            REFERENCES public.cd_workflow(id),
    UNIQUE ("cd_workflow_id"),
    PRIMARY KEY ("id")
);
