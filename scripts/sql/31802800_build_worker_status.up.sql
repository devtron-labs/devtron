CREATE SEQUENCE IF NOT EXISTS id_seq_workflow_execution_stage;


CREATE TABLE public.workflow_execution_stage (
                                                 id int4 NOT NULL DEFAULT nextval('id_seq_workflow_execution_stage'::regclass),
                                                 stage_name varchar(50) NULL,
                                                 status varchar(50) NULL,
                                                 status_type varchar(50) NULL,
                                                 message text NULL,
                                                 metadata text NULL,
                                                 workflow_id int4 NOT NULL,
                                                 workflow_type varchar(50) NOT NULL,
                                                 start_time timestamptz,
                                                 end_time timestamptz,
                                                 created_on timestamptz NOT NULL,
                                                 created_by int4 NOT NULL,
                                                 updated_on timestamptz NOT NULL,
                                                 updated_by int4 NOT null,
                                                 PRIMARY KEY ("id")
);
