CREATE SEQUENCE IF NOT EXISTS id_seq_webhook_config;

CREATE TABLE "public"."webhook_config" (
   "id" integer NOT NULL DEFAULT nextval('id_seq_webhook_config'::regclass),
   "web_hook_url"  VARCHAR(100),
   "config_name"  VARCHAR(100),
   "header"     jsonb,
   "payload"     jsonb,
   "description" text,
   "owner_id"    integer,
   "active"      bool,
   "deleted"     bool NOT NULL DEFAULT FALSE,
   "created_on" timestamptz,
   "created_by" int4,
   "updated_on" timestamptz,
   "updated_by" int4,
   PRIMARY KEY (id)
);