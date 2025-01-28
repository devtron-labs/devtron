BEGIN;

-- Create Sequence for operation_audit
CREATE SEQUENCE IF NOT EXISTS id_seq_operation_audit;

-- Table Definition: operation_audit
CREATE TABLE IF NOT EXISTS "public"."operation_audit" (
    "id"                    int         NOT NULL DEFAULT nextval('id_seq_operation_audit'::regclass),
    "entity_id"             int         NOT NULL,
    "entity_type"           VARCHAR(50) NOT NULL ,
    "operation_type"        VARCHAR(20) NOT NULL,
    "entity_value_json"     jsonb       NOT NULL,
    "schema_for"            VARCHAR(20) NOT NULL,
    "created_on"            timestamptz NOT NULL,
    "created_by"            int4        NOT NULL,
    "updated_on"            timestamptz NOT NULL,
    "updated_by"            int4        NOT NULL,
    PRIMARY KEY ("id")
    );

COMMIT;