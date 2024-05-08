-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_global_tag;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."global_tag"
(
    "id"                        int4         NOT NULL DEFAULT nextval('id_seq_global_tag'::regclass),
    "key"                       varchar(100) NOT NULL,
    "mandatory_project_ids_csv" varchar(100),
    "propagate"                 bool,
    "description"               TEXT         NOT NULL,
    "active"                    bool,
    "created_on"                timestamptz  NOT NULL,
    "created_by"                int4         NOT NULL,
    "updated_on"                timestamptz,
    "updated_by"                int4,
    PRIMARY KEY ("id")
);