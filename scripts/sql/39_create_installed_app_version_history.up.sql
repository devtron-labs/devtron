-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_installed_app_version_history;

-- Table Definition
CREATE TABLE "public"."installed_app_version_history" (
                                                          "id" int4 NOT NULL DEFAULT nextval('id_seq_installed_app_version_history'::regclass),
                                                          "installed_app_version_id" int4 NOT NULL,
                                                          "created_on" timestamptz,
                                                          "created_by" int4,
                                                          "values_yaml_raw" text,
                                                          "status" varchar(100),
                                                          "updated_on" timestamptz,
                                                          "updated_by" int4,
                                                          "git_hash" varchar(255),
                                                          CONSTRAINT "installed_app_version_history_installed_app_version_id_fkey" FOREIGN KEY ("installed_app_version_id") REFERENCES "public"."installed_app_versions"("id"),
                                                          PRIMARY KEY ("id")
);


CREATE INDEX "version_history_git_hash_index" ON "public"."installed_app_version_history" USING BTREE ("git_hash");