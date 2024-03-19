CREATE SEQUENCE IF NOT EXISTS id_seq_timeout_window_resource_mappings;
CREATE TABLE IF NOT EXISTS "public"."timeout_window_resource_mappings"
(
    "id"                int          NOT NULL DEFAULT nextval('id_seq_timeout_window_resource_mappings'::regclass),
    "timeout_window_configuration_id"              int NOT NULL,
    "resource_id" int NOT NULL,
    "resource_type" int NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT timeout_window_configuration_id_fkey
        FOREIGN KEY("timeout_window_configuration_id")
            REFERENCES"public"."timeout_window_configuration" ("id")
            ON DELETE CASCADE
);