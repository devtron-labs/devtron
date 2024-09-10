CREATE SEQUENCE IF NOT EXISTS id_seq_image_tag;

CREATE TABLE IF NOT EXISTS public.release_tags (
    "id"                                         integer NOT NULL DEFAULT nextval('id_seq_image_tag'::regclass),
    "tag_name"                                   varchar(128),
    "artifact_id"                                integer,
    "deleted"                                        BOOL,
    "app_id"                                     integer,
    CONSTRAINT "image_tag_app_id_fkey" FOREIGN KEY ("app_id") REFERENCES "public"."app" ("id"),
    CONSTRAINT "image_tag_artifact_id_fkey" FOREIGN KEY ("artifact_id") REFERENCES "public"."ci_artifact" ("id"),
    UNIQUE ("app_id","tag_name"),
    PRIMARY KEY ("id")
    );

CREATE SEQUENCE IF NOT EXISTS id_seq_image_comment;

CREATE TABLE IF NOT EXISTS public.image_comments (
    "id"                                         integer NOT NULL DEFAULT nextval('id_seq_image_comment'::regclass),
    "comment"                                    varchar(500),
    "artifact_id"                                integer,
    "user_id"                                    integer,
    CONSTRAINT "image_comment_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id"),
    CONSTRAINT "image_comment_artifact_id_fkey" FOREIGN KEY ("artifact_id") REFERENCES "public"."ci_artifact" ("id"),
    PRIMARY KEY ("id")
    );

CREATE SEQUENCE IF NOT EXISTS id_seq_image_tagging_audit;

CREATE TABLE IF NOT EXISTS public.image_tagging_audit (
    "id"                                      integer NOT NULL DEFAULT nextval('id_seq_image_tagging_audit'::regclass),
    "data"                                    text,
    "data_type"                               integer,
    "artifact_id"                             integer,
    "action"                                  integer,
    "updated_on"                              timestamptz,
    "updated_by"                              integer,
    CONSTRAINT "image_tagging_audit_updated_by_fkey" FOREIGN KEY ("updated_by") REFERENCES "public"."users" ("id"),
    CONSTRAINT "image_tagging_audit_artifact_id_fkey" FOREIGN KEY ("artifact_id") REFERENCES "public"."ci_artifact" ("id"),
    PRIMARY KEY ("id")
    );