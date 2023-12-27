CREATE SEQUENCE IF NOT EXISTS id_seq_user_groups;

CREATE TABLE "public"."user_groups"
(
    "id"                        integer          NOT NULL DEFAULT nextval('id_seq_user_groups'::regclass),
    "user_id"                   integer,
    "group_name"                text,
    "is_group_claims_data"      boolean NOT NULL,
    "active"                    boolean NOT NULL,
    "created_on"                timestamptz,
    "created_by"                integer,
    "updated_on"                timestamptz,
    "updated_by"                integer,
    PRIMARY KEY ("id")
);
