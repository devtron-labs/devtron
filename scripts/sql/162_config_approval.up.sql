INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('config', '{ "value": "config", "indexKeyMap": {}}', ARRAY['approve'],'{"value": "%/%/%","indexKeyMap": {"0": "TeamObj","2": "EnvObj","4": "AppObj"}}', ARRAY['apps/devtron-app'],'f','now()', '1', 'now()', '1');


INSERT INTO rbac_role_resource_detail ("resource", "role_resource_key", "role_resource_update_key",
                                       "eligible_entity_access_types", "deleted", "created_on", "created_by",
                                       "updated_on", "updated_by")
VALUES ('approver', 'Approver', 'Approver', ARRAY ['apps/devtron-app'], false, now(), 1, now(), 1);


CREATE SEQUENCE IF NOT EXISTS id_seq_default_rbac_role_data;

CREATE TABLE IF NOT EXISTS "public"."default_rbac_role_data"
(
    "id"                int          NOT NULL DEFAULT nextval('id_seq_default_rbac_role_data'::regclass),
    "role"              varchar(250) NOT NULL,
    "default_role_data" jsonb        NOT NULL,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    "enabled"           boolean      NOT NULL,
    PRIMARY KEY ("id")
);

INSERT INTO "public"."default_rbac_role_data" ( "role","default_role_data", "created_on", "created_by", "updated_on", "updated_by","enabled")
VALUES ('configApprover', '{
    "roleName" : "configApprover",
    "roleDisplayName" : "Config Approver",
    "roleDescription": "Can Approve Draft Config",
    "updatePoliciesForExistingProvidedRoles" : false,
    "entity" : "apps",
    "accessType": "devtron-app",
    "policyResourceList" : [
        {
            "resource": "config",
            "actions" : ["approve"]
        }
    ]
  }', 'now()', '1', 'now()', '1', true);

CREATE SEQUENCE IF NOT EXISTS id_seq_resource_protection;

CREATE TABLE IF NOT EXISTS "public"."resource_protection"
(
    "id"                int NOT NULL DEFAULT nextval('id_seq_resource_protection'::regclass),
    "app_id"            int NOT NULL, /* add foreign key constraint */
    "env_id"            int NOT NULL,
    resource            int NOT NULL,
    protection_state    int NOT NULL,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_resource_protection_history;

CREATE TABLE IF NOT EXISTS "public"."resource_protection_history"
(
    "id"                int NOT NULL DEFAULT nextval('id_seq_resource_protection_history'::regclass),
    app_id              int NOT NULL,
    env_id              int NOT NULL,
    resource            int NOT NULL,
    protection_state    int NOT NULL,
    updated_on          timestamptz,
    updated_by          integer,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_draft;

CREATE TABLE IF NOT EXISTS "public"."draft"
(
    id                int     NOT NULL DEFAULT nextval('id_seq_draft'::regclass),
    app_id            int     NOT NULL, /* add foreign key constraint */
    env_id            int     NOT NULL,
    resource          int     NOT NULL,
    resource_name     varchar(300)     NOT NULL,
    draft_state       int,
    created_on        timestamptz,
    created_by        integer,
    updated_on        timestamptz,
    updated_by        integer,
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_draft_version;

CREATE TABLE IF NOT EXISTS "public"."draft_version"
(
    id                int     NOT NULL DEFAULT nextval('id_seq_draft_version'::regclass),
    draft_id          int     NOT NULL,
    data              text    NOT NULL,
    action            int     NOT NULL,
    user_id           int     NOT NULL,
    created_on        timestamptz,
    CONSTRAINT "drafts_relation_draft_id_fkey" FOREIGN KEY ("draft_id") REFERENCES "public"."draft" ("id"),
    PRIMARY KEY ("id")
);

CREATE SEQUENCE IF NOT EXISTS id_seq_draft_version_comment;

CREATE TABLE IF NOT EXISTS "public"."draft_version_comment"
(
    id                int     NOT NULL DEFAULT nextval('id_seq_draft_version_comment'::regclass),
    draft_id          int     NOT NULL,
    draft_version_id  int     NOT NULL,
    comment           text,
    active            bool     NOT NULL,
    created_on        timestamptz,
    created_by        integer,
    updated_on        timestamptz,
    updated_by        integer,
    CONSTRAINT "drafts_relation_draft_id_fkey" FOREIGN KEY ("draft_id") REFERENCES "public"."draft" ("id"),
    CONSTRAINT "draft_versions_relation_draft_version_id_fkey" FOREIGN KEY ("draft_version_id") REFERENCES "public"."draft_version" ("id"),
    PRIMARY KEY ("id")
);
