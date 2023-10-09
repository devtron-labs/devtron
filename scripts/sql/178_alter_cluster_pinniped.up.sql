ALTER TABLE cluster
    ADD COLUMN to_connect_with_pinniped boolean,
    ADD COLUMN pinniped_concierge_url   text;


CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_tool_data;

CREATE TABLE "public"."cluster_tool_data"
(
    "id"         integer      NOT NULL DEFAULT nextval('id_seq_cluster_tool_data'::regclass),
    "cluster_id" integer      NOT NULL,
    "user_id"    integer      NOT NULL,
    "tool_name"  varchar(250) NOT NULL,
    "token"      text         NOT NULL,
    CONSTRAINT "cluster_tool_data_cluster_id_fkey" FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster" ("id"),
    CONSTRAINT "cluster_tool_data_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id"),
    PRIMARY KEY ("id")
);
