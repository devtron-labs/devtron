-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_api_token;

-- Table Definition
CREATE TABLE "public"."api_token"
(
    "id"              int4        NOT NULL DEFAULT nextval('id_seq_api_token'::regclass),
    "user_id"         int4        NOT NULL,
    "name"            varchar(50) NOT NULL UNIQUE,
    "description"     text        NOT NULL,
    "expire_at_in_ms" bigint, -- null means never
    "token"           text        NOT NULL UNIQUE,
    "created_on"      timestamptz NOT NULL,
    "created_by"      int4,
    "updated_on"      timestamptz,
    "updated_by"      int4,
    PRIMARY KEY ("id")
);

-- add foreign key
ALTER TABLE "public"."api_token" ADD FOREIGN KEY ("user_id") REFERENCES "public"."users"("id");

-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_user_audit;

-- Table Definition
CREATE TABLE "public"."user_audit"
(
    "id"         int4         NOT NULL DEFAULT nextval('id_seq_user_audit'::regclass),
    "user_id"    int4         NOT NULL,
    "client_ip"  varchar(256) NOT NULL,
    "created_on" timestamptz  NOT NULL,
    PRIMARY KEY ("id")
);

-- add foreign key
ALTER TABLE "public"."user_audit" ADD FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id");

--- Create index on user_audit.user_id
CREATE INDEX user_audit_user_id_IX ON public.user_audit (user_id);

-- insert secret into attributes table
INSERT INTO attributes(key, value, active, created_on, created_by)
VALUES ('apiTokenSecret', MD5(random()::text), 't', NOW(), 1);

-- add column user_type in user table
ALTER TABLE users ADD COLUMN user_type varchar(250);