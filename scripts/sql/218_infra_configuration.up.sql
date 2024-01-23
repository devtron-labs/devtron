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
    PRIMARY KEY ("id")
    FOREIGN KEY ("profile_id") REFERENCES build_infra_profile ("id")
    );

