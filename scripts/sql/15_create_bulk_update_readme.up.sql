CREATE TABLE "public"."bulk_update_readme"
(
    "id"        int4         NOT NULL,
    "operation" varchar(255) NOT NULL,
    "readme"    text,
    "script"    jsonb,
    PRIMARY KEY ("id")
);