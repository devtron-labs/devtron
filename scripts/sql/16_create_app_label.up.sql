CREATE SEQUENCE IF NOT EXISTS id_seq_app_label;

-- Table Definition
CREATE TABLE "public"."app_label"
(
    "id"         int4 NOT NULL DEFAULT nextval('id_seq_app_label'::regclass),
    "app_id"     int4,
    "key"        varchar(255),
    "value"      varchar(255),
    "created_on" timestamptz,
    "created_by" int4,
    "updated_on" timestamptz,
    "updated_by" int4,
    CONSTRAINT "app_label_app_id_fkey" FOREIGN KEY ("app_id") REFERENCES "public"."app" ("id"),
    PRIMARY KEY ("id")
);

CREATE UNIQUE INDEX "app_label_unique" ON "public"."app_label" USING BTREE ("app_id","key");