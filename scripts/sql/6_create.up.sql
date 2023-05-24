-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_gitops_config;

-- Table Definition
CREATE TABLE "public"."gitops_config" (
    "id" int4 NOT NULL DEFAULT nextval('id_seq_gitops_config'::regclass),
    "provider" varchar(250) NOT NULL,
    "username" varchar(250) NOT NULL,
    "token" varchar(250) NOT NULL,
    "github_org_id" varchar(250),
    "host" varchar(250) NOT NULL,
    "active" bool NOT NULL,
    "created_on" timestamptz,
    "created_by" integer,
    "updated_on" timestamptz,
    "updated_by" integer,
    "gitlab_group_id" varchar(250),
    PRIMARY KEY ("id")
);