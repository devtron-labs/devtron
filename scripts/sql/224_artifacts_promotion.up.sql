ALTER TABLE deployment_approval_user_data ADD COLUMN "resource_type" integer DEFAULT 0;
-- 0 for  deployment approval request, 1 for artifact promotion approval request
ALTER TABLE deployment_approval_user_data RENAME COLUMN "approval_request_id" TO "resource_approval_request_id";
-- rename deployment_approval_user_data table to resource_approval_user_data
ALTER TABLE deployment_approval_user_data RENAME TO resource_approval_user_data;

ALTER TABLE  resource_filter_evaluation_audit ADD COLUMN "resource_type" integer DEFAULT 0;
-- 0 for  resource_filter, 1 for artifact promotion policy filter evaluation

-- create artifact promotion policy table
CREATE SEQUENCE IF NOT EXISTS id_artifact_promotion_policy;
CREATE TABLE IF NOT EXISTS public.artifact_promotion_policy
(
    "active"                       bool         NOT NULL,
    "created_by"                   int4         NOT NULL,
    "updated_by"                   int4         NOT NULL,
    "id"                           int          NOT NULL DEFAULT nextval('id_artifact_promotion_policy'::regclass),
    approval_count                 int          NOT NULL,
    "name"                         VARCHAR(50)  NOT NULL,
    "description"                  VARCHAR(300),
    "created_on"                   timestamptz  NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "condition_expression"         text         NOT NULL,
--  this column contains {"restrictImageBuilderFromApprove": false, "restrictPromoterFromApprove": false, "restrictApproverFromDeploy": false}
    "approval_metadata"            json         NOT NULL,

    PRIMARY KEY ("id")
    );
CREATE UNIQUE INDEX idx_unique_promotion_policy_name
    ON artifact_promotion_policy(name)
    WHERE active = true;

-- promotion policies audit table, stores the auditing for delete,create and update actions
CREATE SEQUENCE IF NOT EXISTS artifact_promotion_policy_audit_seq;
CREATE TABLE IF NOT EXISTS "public"."artifact_promotion_policy_audit"
(
    "id"  integer not null default nextval('resource_filter_audit_seq' :: regclass),
    "policy_data" text     NOT NULL,
    "policy_id"   int      NOT NULL,
--     action is either create, update ,delete
    "action"      int      NOT NULL,
    "created_on"        timestamptz,
    "created_by"        integer,
    "updated_on"        timestamptz,
    "updated_by"        integer,
    CONSTRAINT "artifact_promotion_policy_audit_policy_id_fkey" FOREIGN KEY ("policy_id") REFERENCES "public"."artifact_promotion_policy" ("id"),
    PRIMARY KEY ("id")
    );

-- create artifact promotion approval request table
CREATE SEQUENCE IF NOT EXISTS id_artifact_promotion_approval_request;
CREATE TABLE IF NOT EXISTS public.artifact_promotion_approval_request
(
    "active"                       bool         NOT NULL,
    --     foreign key to user, promoted_by
    "requested_by"                 int4         NOT NULL,
    "id"                           int          NOT NULL DEFAULT nextval('id_artifact_promotion_approval_request'::regclass),
--     foreign key to artifact_promotion_policy
    "policy_id"                    int          NOT NULL,
--     foreign key to filter_evaluation_audit
    "policy_evaluation_audit_id"   int          NOT NULL
--     foreign key to ci_artifact
    "artifact_id"                  int          NOT NULL,
    "source_pipeline_id"           int          NOT NULL,
--     CI_PIPELINE(0) or WEBHOOK(1) or CD_PIPELINE(2)
    "source_type"                  int          NOT NULL,
    "destination_pipeline_id"      int          NOT NULL,
--  CD_PIPELINE(2) , currently not defining this column as destination is always CD_PIPELINE
--  "destination_type"             int          NOT NULL,
    "status"                       int          NOT NULL,
--  promoted_on time
    "promoted_on"                  timestamptz  NOT NULL,
--  promoted_on time
    "requested_on"                 timestamptz  NOT NULL,

    PRIMARY KEY ("id")
    );

CONSTRAINT "artifact_promotion_approval_request_policy_id_fkey" FOREIGN KEY ("policy_id") REFERENCES "public"."artifact_promotion_policy" ("id");
CONSTRAINT "artifact_promotion_approval_request_artifact_id_fkey" FOREIGN KEY ("artifact_id") REFERENCES "public"."ci_artifact" ("id");
CONSTRAINT "artifact_promotion_approval_request_policy_evaluation_audit_id_fkey" FOREIGN KEY ("policy_evaluation_audit_id") REFERENCES "public"."resource_filter_evaluation_audit" ("id");
CREATE UNIQUE INDEX "idx_unique_artifact_promoted_to_destination"
    ON artifact_promotion_approval_request(artifact_id,destination_pipeline_id)
    WHERE status = 'PROMOTED';

-- custom role queries
insert into rbac_policy_resource_detail
(resource,
 policy_resource_value,
 allowed_actions,
 resource_object,
 eligible_entity_access_types,
 deleted,created_on,
 created_by,
 updated_on,
 updated_by)
values ('config/artifact',
        '{"value": "config/artifact", "indexKeyMap": {}}','{promoter}','{"value": "%/%/%", "indexKeyMap": {"0": "TeamObj", "2": "EnvObj", "4": "AppObj"}}','{apps/devtron-app}',
        false,
        now(),
        1,
        now(),
        1);



insert into default_rbac_role_data (role,
                                    default_role_data,
                                    created_on,
                                    created_by,
                                    updated_on,
                                    updated_by,
                                    enabled)
VALUES ('artifactPromoter',
        '{"entity": "apps", "roleName": "artifactPromoter", "accessType": "devtron-app", "roleDescription": "can promote artifact for a particular CD Pipeline", "roleDisplayName": "Artifact Promoter", "policyResourceList": [{"actions": ["promoter"],
"resource": "config/artifact"}], "updatePoliciesForExistingProvidedRoles": false}',
        now(),
        1,
        now(),
        1,
        true);