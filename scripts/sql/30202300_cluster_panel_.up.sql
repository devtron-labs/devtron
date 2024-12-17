-- File: scripts/sql/30202300_cluster_panel_.up.sql

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_panel;

-- Table Definition
CREATE TABLE "public"."panel" (
                                  "id" int4 NOT NULL DEFAULT nextval('id_seq_panel'::regclass),
                                  "name" varchar(250) NOT NULL,
                                  "cluster_id" int4 NOT NULL,
                                  "active" bool NOT NULL,
                                  "embed_iframe" text,
                                  "created_on" timestamptz,
                                  "created_by" integer,
                                  "updated_on" timestamptz,
                                  "updated_by" integer,
                                  PRIMARY KEY ("id")
);