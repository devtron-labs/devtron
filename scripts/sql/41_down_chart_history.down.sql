CREATE TABLE "public"."app_store_charts_history"
(
    "id"                            integer NOT NULL DEFAULT nextval('id_seq_app_store_charts_history'::regclass),
    "installed_apps_id"             integer NOT NULL,
    "values_yaml"                   text,
    "deployed_on"                   timestamptz,
    "deployed_by"                   int4,
    "created_on"                    timestamptz,
    "created_by"                    int4,
    "updated_on"                    timestamptz,
    "updated_by"                    int4,
    CONSTRAINT "app_store_charts_history_installed_apps_id_fkey" FOREIGN KEY ("installed_apps_id") REFERENCES "public"."installed_apps" ("id"),
    PRIMARY KEY ("id")
);