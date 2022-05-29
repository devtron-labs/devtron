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
    "last_used_at"    timestamptz,
    "last_used_by_ip" varchar(50),
    "created_on"      timestamptz NOT NULL,
    "created_by"      int4,
    "updated_on"      timestamptz,
    "updated_by"      int4,
    CONSTRAINT "api_token_user_id_fkey" FOREIGN KEY ("user_id") REFERENCES "public"."users" ("id"),
    PRIMARY KEY ("id")
);


-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_api_token_secret;

-- Table Definition
CREATE TABLE "public"."api_token_secret"
(
    "id"         int4         NOT NULL DEFAULT nextval('id_seq_api_token_secret'::regclass),
    "secret"     varchar(250) NOT NULL,
    "created_on" timestamptz  NOT NULL,
    "updated_on" timestamptz,
    PRIMARY KEY ("id")
);

---- Insert master data into api_token_secret
INSERT INTO api_token_secret (secret, created_on)
VALUES (MD5(random()::text), NOW());