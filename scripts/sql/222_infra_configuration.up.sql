CREATE SEQUENCE IF NOT EXISTS id_seq_infra_profile;
CREATE TABLE IF NOT EXISTS public.infra_profile
(
    "id"                           int          NOT NULL DEFAULT nextval('id_seq_infra_profile'::regclass),
    "name"                         VARCHAR(50)  UNIQUE NOT NULL,
    "description"                  VARCHAR(300),
    "active"                       bool         NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id")
    );

<<<<<<<< HEAD:scripts/sql/222_infra_configuration.up.sql
CREATE UNIQUE INDEX idx_unique_name
========
CREATE UNIQUE INDEX idx_unique_profile_name
>>>>>>>> build-infra-profile:scripts/sql/222_infra_configuration.up.sql
    ON infra_profile (name)
    WHERE active = true;

CREATE SEQUENCE IF NOT EXISTS id_seq_infra_profile_configuration;

CREATE TABLE IF NOT EXISTS public.infra_profile_configuration
(
    "id"                           int          NOT NULL DEFAULT nextval('id_seq_infra_profile_configuration'::regclass),
    "key"                          int          NOT NULL,
    "value"                        float        NOT NULL,
    "profile_id"                   int          NOT NULL,
    "unit"                         int          NOT NULL,
    "active"                       bool         NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "updated_by"                   int4         NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "infra_profile_configuration_profile_id_fkey" FOREIGN KEY ("profile_id") REFERENCES "public"."infra_profile" ("id")
    );

