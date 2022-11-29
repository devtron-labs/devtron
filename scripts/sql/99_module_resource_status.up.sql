-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_module_resource_status;

-- Table Definition
CREATE TABLE "public"."module_resource_status"
(
    "id"             int4         NOT NULL DEFAULT nextval('id_seq_module_resource_status'::regclass),
    "module_id"      int4         NOT NULL,
    "group"          varchar(50)  NOT NULL,
    "version"        varchar(50)  NOT NULL,
    "kind"           varchar(50)  NOT NULL,
    "name"           varchar(250) NOT NULL,
    "health_status"  varchar(50),
    "health_message" varchar(1024),
    "active"         bool,
    "created_on"     timestamptz  NOT NULL,
    "updated_on"     timestamptz,
    PRIMARY KEY ("id")
);

-- add foreign key
ALTER TABLE "public"."module_resource_status"
    ADD FOREIGN KEY ("module_id") REFERENCES "public"."module" ("id");

