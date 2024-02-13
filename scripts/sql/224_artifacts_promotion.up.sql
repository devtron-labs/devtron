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
    "name"                         VARCHAR(50)  NOT NULL,
    "description"                  VARCHAR(300),
    "created_on"                   timestamptz  NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    "condition_expression"         text         NOT NULL,

    PRIMARY KEY ("id")
    );
CREATE UNIQUE INDEX idx_unique_promotion_policy_name
    ON artifact_promotion_policy(name)
    WHERE active = true;

-- create artifact promotion approval request table
CREATE SEQUENCE IF NOT EXISTS id_artifact_promotion_approval_request;
CREATE TABLE IF NOT EXISTS public.artifact_promotion_approval_request
(
    "active"                       bool         NOT NULL,
    --     foreign key to user
    "created_by"                   int4         NOT NULL,
    --     foreign key to user
    "updated_by"                   int4         NOT NULL,
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
--     CD_PIPELINE(2)
    "destination_type"             int          NOT NULL,
    "status"                       int          NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,

    PRIMARY KEY ("id")
    );

CONSTRAINT "artifact_promotion_approval_request_policy_id_fkey" FOREIGN KEY ("policy_id") REFERENCES "public"."artifact_promotion_policy" ("id");
CONSTRAINT "artifact_promotion_approval_request_artifact_id_fkey" FOREIGN KEY ("artifact_id") REFERENCES "public"."ci_artifact" ("id");
CONSTRAINT "artifact_promotion_approval_request_policy_evaluation_audit_id_fkey" FOREIGN KEY ("policy_evaluation_audit_id") REFERENCES "public"."resource_filter_evaluation_audit" ("id");
