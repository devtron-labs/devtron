CREATE SEQUENCE IF NOT EXISTS "public"."id_seq_sso_login_config";

CREATE TABLE "public"."sso_login_config"
  (
     "id"         INT4 NOT NULL DEFAULT NEXTVAL('id_seq_sso_login_config'::
     regclass),
     "name"       VARCHAR(250),
     "label"      VARCHAR(250),
     "url"        VARCHAR(250),
     "config"     TEXT,
     "created_on" TIMESTAMPTZ,
     "created_by" INT4,
     "updated_on" TIMESTAMPTZ,
     "updated_by" INT4,
     "active"     BOOL,
     PRIMARY KEY ("id")
  );