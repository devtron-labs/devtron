BEGIN;

-- Create Sequence for infra_config_trigger_history
CREATE SEQUENCE IF NOT EXISTS "public"."id_seq_infra_config_trigger_history";

-- Table Definition: infra_config_trigger_history
CREATE TABLE IF NOT EXISTS "public"."infra_config_trigger_history" (
    "id"                    int             NOT NULL DEFAULT nextval('id_seq_infra_config_trigger_history'::regclass),
    "key"                   int             NOT NULL,
    "value_string"          text,
    "platform"              varchar(50)     NOT NULL,
    "workflow_id"           int             NOT NULL,
    "workflow_type"         varchar(255)    NOT NULL,
    "created_on"            timestamptz     NOT NULL,
    "created_by"            int4            NOT NULL,
    "updated_on"            timestamptz     NOT NULL,
    "updated_by"            int4            NOT NULL,
    PRIMARY KEY ("id"),
    UNIQUE ("workflow_id", "workflow_type", "key", "platform")
);

END;