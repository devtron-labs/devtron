CREATE SEQUENCE IF NOT EXISTS id_seq_deployment_app_migration_history;
CREATE TABLE "public"."deployment_app_migration_history" (
	  "id" integer NOT NULL default nextval('id_seq_deployment_app_migration_history'::regclass),
	  "app_id" integer,
      "env_id" integer,
	  "is_migration_active" bool NOT NULL,
      "migrate_to" text ,
      "migrate_from" text,
      "current_status" integer,
      "error_status" integer,
      "error_encountered" text,
	  "created_on"                timestamptz NOT NULL,
	  "created_by"                int4        NOT NULL,
	  "updated_on"                timestamptz,
	  "updated_by"                int4,
	  PRIMARY KEY ("id"),
      CONSTRAINT "deployment_app_migration_history_app_id_fk" FOREIGN KEY ("app_id") REFERENCES "public"."app" ("id"),
      CONSTRAINT "deployment_app_migration_history_env_id_fk" FOREIGN KEY ("env_id") REFERENCES "public"."environment" ("id")
);