CREATE SEQUENCE IF NOT EXISTS id_seq_resource_qualifier_mapping_criteria;

CREATE TABLE IF NOT EXISTS resource_qualifier_mapping_criteria
(
    "id"         integer     NOT NULL DEFAULT nextval('id_seq_resource_qualifier_mapping_criteria'::regclass),
    "description"       varchar(100),
    "json_data" text,
    "active" bool,
    "created_on" timestamptz NOT NULL,
    "created_by" int4        NOT NULL,
    "updated_on" timestamptz NOT NULL,
    "updated_by" int4        NOT NULL,
    PRIMARY KEY ("id")
);

INSERT INTO devtron_resource_searchable_key(name, is_removed, created_on, created_by, updated_on, updated_by)
VALUES ('BASE_DEPLOYMENT_TEMPLATE', false, now(), 1, now(), 1);

INSERT INTO devtron_resource_searchable_key(name, is_removed, created_on, created_by, updated_on, updated_by)
VALUES ('GLOBAL_ID', false, now(), 1, now(), 1);