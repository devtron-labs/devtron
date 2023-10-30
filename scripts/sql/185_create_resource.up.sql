-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_devtron_resource;

-- Table Definition
CREATE TABLE "public"."devtron_resource"
(
    "id"             int          NOT NULL DEFAULT nextval('id_seq_devtron_resource'::regclass),
    "kind"           varchar(250) NOT NULL,
    "parent_kind_id" int,
    "deleted"        boolean,
    "created_on"     timestamptz,
    "created_by"     integer,
    "updated_on"     timestamptz,
    "updated_by"     integer,
    PRIMARY KEY ("id")
);
