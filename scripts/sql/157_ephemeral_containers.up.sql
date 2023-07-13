CREATE SEQUENCE IF NOT EXISTS id_seq_ephemeral_container;

CREATE TABLE "public"."ephemeral_container" (
    "id"                      integer NOT NULL DEFAULT nextval('id_seq_ephemeral_container'::regclass),
    "name"                    VARCHAR(253) NOT NULL,
    "cluster_id"              INTEGER NOT NULL,
    "namespace"               VARCHAR(250) NOT NULL,
    "pod_name"                VARCHAR(250) NOT NULL,
    "target_container"        VARCHAR(250) NOT NULL,
    "config"                  TEXT NOT NULL,
    "is_externally_created"   BOOLEAN NOT NULL DEFAULT FALSE,
    CONSTRAINT "ephemeral_container_cluster_id_fkey" FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_ephemeral_container_actions;

CREATE TABLE "public"."ephemeral_container_actions" (
    "id"                     INTEGER NOT NULL DEFAULT nextval('id_seq_ephemeral_container_actions'::regclass),
    "ephemeral_container_id" INTEGER NOT NULL,
    "action_type"            INTEGER DEFAULT 0,
    "performed_by"           INTEGER NOT NULL,
    "performed_at"           timestamptz NOT NULL,
    CONSTRAINT "ephemeral_container_actions_ephemeral_container_id_fkey" FOREIGN KEY ("ephemeral_container_id") REFERENCES "public"."ephemeral_container" ("id"),
    CONSTRAINT "ephemeral_container_actions_performed_by_fkey" FOREIGN KEY ("performed_by") REFERENCES "public"."users" ("id"),
    PRIMARY KEY (id)
);

