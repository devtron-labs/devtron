CREATE SEQUENCE IF NOT EXISTS id_seq_ci_template_override;

-- Table Definition
CREATE TABLE "public"."ci_template_override"
(
    "id"                          integer NOT NULL DEFAULT nextval('id_seq_ci_template_override'::regclass),
    "ci_pipeline_id"              integer,
    "docker_registry_id"          text,
    "docker_repository"           text,
    "dockerfile_path"              text,
    "git_material_id"             integer,
    "active"                      boolean,
    "created_on"                  timestamptz,
    "created_by"                  int4,
    "updated_on"                  timestamptz,
    "updated_by"                  int4,
    CONSTRAINT "ci_template_override_ci_pipeline_id_fkey" FOREIGN KEY ("ci_pipeline_id") REFERENCES "public"."ci_pipeline" ("id"),
    CONSTRAINT "ci_template_override_git_material_id_fkey" FOREIGN KEY ("git_material_id") REFERENCES "public"."git_material" ("id"),
    PRIMARY KEY ("id")
);


ALTER TABLE "public"."ci_pipeline" ADD COLUMN "is_docker_config_overridden" boolean DEFAULT FALSE;