CREATE SEQUENCE IF NOT EXISTS id_seq_timeout_window_configuration;


-- Table Definition
CREATE TABLE "public"."timeout_window_configuration" (
    "id" int NOT NULL DEFAULT nextval('id_seq_timeout_window_configuration'::regclass),
    "timeout_window_expression" varchar(255) NOT NULL,
    "timeout_window_expression_format" int NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
     PRIMARY KEY ("id")
);


ALTER TABLE "public"."users"
    ADD COLUMN "timeout_window_configuration_id" int;

ALTER TABLE "public"."users" ADD FOREIGN KEY ("timeout_window_configuration_id")
    REFERENCES "public"."timeout_window_configuration" ("id");