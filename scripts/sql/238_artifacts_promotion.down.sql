ALTER TABLE request_approval_user_data DROP COLUMN "request_type";

ALTER TABLE request_approval_user_data RENAME TO deployment_approval_user_data;

ALTER TABLE deployment_approval_user_data
    ADD CONSTRAINT deployment_approval_user_data_approval_request_id_fkey
        FOREIGN KEY ("approval_request_id")
            REFERENCES "public"."deployment_approval_request" ("id");

DROP INDEX unique_user_request_action;

ALTER TABLE "resource_filter_evaluation_audit" DROP COLUMN "filter_type";

DROP INDEX IF EXISTS unique_user_request_action;

ALTER TABLE "resource_filter_evaluation_audit" DROP COLUMN "filter_type";

DROP SEQUENCE IF EXISTS id_artifact_promotion_approval_request;

DROP TABLE "public"."artifact_promotion_approval_request";

DROP INDEX idx_unique_artifact_promoted_to_destination;

DELETE from rbac_policy_resource_detail where resource = 'artifact';

DELETE from default_rbac_role_data where role='artifactPromoter';

DELETE from event where id=7;

DELETE from notification_templates where event_type_id=7;








