BEGIN;
-- Sequence
CREATE SEQUENCE IF NOT EXISTS id_seq_license_attributes;

-- Table Definition
CREATE TABLE IF NOT EXISTS "public"."license_attributes" (
                                       "id" int4 NOT NULL DEFAULT nextval('id_seq_license_attributes'::regclass),
                                       "key" varchar(250) NOT NULL,
                                       "value" TEXT NOT NULL,
                                       "active" bool NOT NULL,
                                       "created_on" timestamptz,
                                       "created_by" integer,
                                       "updated_on" timestamptz,
                                       "updated_by" integer,
                                       PRIMARY KEY ("id")
);

COMMIT;