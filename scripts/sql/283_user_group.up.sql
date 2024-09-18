CREATE SEQUENCE IF NOT EXISTS id_seq_user_group;
CREATE TABLE IF NOT EXISTS public.user_group
(
    "id"                           int          NOT NULL DEFAULT nextval('id_seq_user_group'::regclass),
    "name"                         VARCHAR(50)  NOT NULL,
    "identifier"                   VARCHAR(50)  NOT NULL,
    "description"                  TEXT         NOT NULL,
    "active"                       bool         NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id")
    );

CREATE UNIQUE INDEX idx_unique_user_group_name
    ON user_group (name)
    WHERE active = true;

CREATE UNIQUE INDEX idx_unique_user_group_identifier
    ON user_group (identifier)
    WHERE active = true;

CREATE SEQUENCE IF NOT EXISTS id_seq_user_group_mapping;
CREATE TABLE IF NOT EXISTS public.user_group_mapping
(
    "id"                           int          NOT NULL DEFAULT nextval('id_seq_user_group_mapping'::regclass),
    "user_id"                      int          NOT NULL,
    "user_group_id"                int          NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "user_group_mapping_user_group_id_fkey" FOREIGN KEY ("user_group_id") REFERENCES "public"."user_group" ("id"),
    CONSTRAINT "user_group_mapping_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id")
    );

CREATE UNIQUE INDEX idx_unique_user_group_user_id
    ON user_group_mapping(user_id,user_group_id);

ALTER TABLE user_groups RENAME TO user_auto_assigned_groups;
