CREATE SEQUENCE IF NOT EXISTS id_seq_ci_env_mapping;
CREATE TABLE "public"."ci_env_mapping" (
"id" integer NOT NULL DEFAULT nextval('id_seq_ci_env_mapping'::regclass),
"environment_id" integer,
"ci_pipeline_id" integer,
"deleted"     bool NOT NULL DEFAULT FALSE,
"created_on" timestamptz,
"created_by" int4,
"updated_on" timestamptz,
"updated_by" int4,
CONSTRAINT "ci_env_mapping_ci_pipeline_id_fkey" FOREIGN KEY ("ci_pipeline_id") REFERENCES "public"."ci_pipeline" ("id"),
CONSTRAINT "ci_env_mapping_environment_id_fkey" FOREIGN KEY ("environment_id") REFERENCES "public"."environment" ("id"),
PRIMARY KEY (id)
);