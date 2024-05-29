/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- 1 for  deployment approval request, 2 for artifact promotion approval request
ALTER TABLE deployment_approval_user_data ADD COLUMN "request_type" integer NOT NULL DEFAULT 1;

--  drop the constraint as this is no longer valid
ALTER TABLE deployment_approval_user_data DROP CONSTRAINT deployment_approval_user_data_approval_request_id_fkey;
ALTER TABLE deployment_approval_user_data DROP CONSTRAINT deployment_approval_user_data_approval_request_id_user_id_key;

-- rename deployment_approval_user_data table to request_approval_user_data
ALTER TABLE deployment_approval_user_data RENAME TO request_approval_user_data;

-- user can take action only once on any approval_request
CREATE UNIQUE INDEX "unique_user_request_action"
    ON request_approval_user_data(user_id,approval_request_id,request_type);
-- 1 for  resource_filter, 2 for artifact promotion policy filter evaluation


-- create artifact promotion approval request table
CREATE SEQUENCE IF NOT EXISTS id_artifact_promotion_approval_request;
CREATE TABLE IF NOT EXISTS public.artifact_promotion_approval_request
(
    "created_by"                   int4         NOT NULL,
    "updated_by"                   int4         NOT NULL,
    "id"                           int          NOT NULL DEFAULT nextval('id_artifact_promotion_approval_request'::regclass),
--     foreign key to artifact_promotion_policy
    "policy_id"                    int          NOT NULL,
--     foreign key to filter_evaluation_audit
    "policy_evaluation_audit_id"   int          NOT NULL,
--     foreign key to ci_artifact
    "artifact_id"                  int          NOT NULL,
    "source_pipeline_id"           int          NOT NULL,
--     CI_PIPELINE(0) or WEBHOOK(1) or CD_PIPELINE(2)
    "source_type"                  int          NOT NULL,
    "destination_pipeline_id"      int          NOT NULL,
--  CD_PIPELINE(2) , currently not defining this column as destination is always CD_PIPELINE
--  "destination_type"             int          NOT NULL,
    "status"                       int          NOT NULL,
    "created_on"                   timestamptz  NOT NULL,
    "updated_on"                   timestamptz  NOT NULL,
    PRIMARY KEY ("id"),
    CONSTRAINT "artifact_promotion_approval_request_policy_id_fkey" FOREIGN KEY ("policy_id") REFERENCES "public"."global_policy" ("id"),
    CONSTRAINT "artifact_promotion_approval_request_artifact_id_fkey" FOREIGN KEY ("artifact_id") REFERENCES "public"."ci_artifact" ("id"),
    CONSTRAINT "artifact_promotion_approval_request_policy_evaluation_audit_id_fkey" FOREIGN KEY ("policy_evaluation_audit_id") REFERENCES "public"."resource_filter_evaluation_audit" ("id")
    );


CREATE UNIQUE INDEX "idx_unique_artifact_promoted_to_destination"
    ON artifact_promotion_approval_request(artifact_id,destination_pipeline_id)
    WHERE (status = 3 and status = 1);

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
values ('artifact',
        '{"value": "artifact", "indexKeyMap": {}}','{promote}','{"value": "%/%/%", "indexKeyMap": {"0": "TeamObj", "2": "EnvObj", "4": "AppObj"}}','{apps/devtron-app}',
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
        '{"entity": "apps", "roleName": "artifactPromoter", "accessType": "devtron-app", "roleDescription": "can promote artifact for a particular CD Pipeline", "roleDisplayName": "Artifact Promoter", "policyResourceList": [{"actions": ["promote"],
"resource": "artifact"}], "updatePoliciesForExistingProvidedRoles": false}',
        now(),
        1,
        now(),
        1,
        true);


INSERT INTO public.event (id, event_type, description) VALUES (7, 'PROMOTION APPROVAL', '');


INSERT INTO "public"."notification_templates" (channel_type, node_type, event_type_id, template_name, template_payload)
VALUES ('smtp', 'CD', 7, 'image promotion smtp template', '{
    "from": "{{fromEmail}}",
    "to": "{{toEmail}}",
    "subject": "🛎️ Image Promotion Approval Requested | Application: {{appName}} | Target environment: {{envName}}",
    "html": "<table cellpadding=\"0\" style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=\"3\"><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img style=\"max-width:122px\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\"></div></td></tr><tr><td colspan=\"3\"><div style=\"background-color:#e5f2ff;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=\"width:90%\"><div style=\"font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14\">Image promotion approval request</div><span style=\"font-size:14px;line-height:20px;color:#000a14\">{{eventTime}}</span><br><div><span style=\"font-size:14px;line-height:20px;color:#000a14\">by</span><span style=\"font-size:14px;line-height:20px;color:#06c;margin-left:4px\">{{triggeredBy}}</span></div></div><div><img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height:72px;width:72px\"></div></div></td></tr><tr><td colspan=\"3\"><div style=\"display:flex\"><div style=\"background-color:#e5f2ff;border-bottom-left-radius:8px;padding:0 0 20px 20px\"><a href=\"{{&artifactPromotionApprovalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">Approve</a></div><div style=\"width:90%;background-color:#e5f2ff;border-bottom-right-radius:8px;padding:0 0 20px 8px\"><a href=\"{{&artifactPromotionRequestViewLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer;background-color:#fff\">View Request</a></div></div></td></tr><tr></tr><tr><td><br></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px\">Application</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{appName}}</div></td></tr><tr><td><br></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px\">Source</div></td><td colspan=\"2\"><div style=\"color:#3b444c;font-size:13px\">Target environment</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{promotionArtifactSource}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{envName}}</div></td></tr><tr><td colspan=\"3\"><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Image Details</div></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px\">Image tag</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{imageTag}}</div></td></tr><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display:{{commentDisplayStyle}};\">Comment</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display:{{commentDisplayStyle}};\">{{comment}}</div></td></tr><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display:{{tagDisplayStyle}};\">Tags</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display:{};\">{{tags}}</div></td></tr><tr><td colspan=\"3\"><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div></td></tr><tr><td colspan=\"2\" style=\"display:flex\"><span><a href=\"https://twitter.com/DevtronL\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/twitter_social_dark.png\"></div></a></span><span><a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"></div></a></span><span><a href=\"https://devtron.ai/blog/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px\">Blog</a></span><span><a href=\"https://devtron.ai/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline\">Website</a></span></td><td colspan=\"2\" style=\"text-align:right\"><div style=\"color:#767d84;font-size:13px;line-height:20px\">&copy; Devtron Labs 2024</div></td></tr></table>"
}');

INSERT INTO "public"."notification_templates" (channel_type, node_type, event_type_id, template_name, template_payload)
VALUES ('ses', 'CD', 7, 'image promotion ses template', '{
    "from": "{{fromEmail}}",
    "to": "{{toEmail}}",
    "subject": "🛎️ Image Promotion Approval Requested | Application: {{appName}} | Target environment: {{envName}}",
    "html": "<table cellpadding=\"0\" style=\"font-family:Arial,Verdana,Helvetica;width:600px;height:485px;border-collapse:inherit;border-spacing:0;border:1px solid #d0d4d9;border-radius:8px;padding:16px 20px;margin:20px auto;box-shadow:0 0 8px 0 rgba(0,0,0,.1)\"><tr><td colspan=\"3\"><div style=\"padding-bottom:16px;margin-bottom:20px;border-bottom:1px solid #edf1f5;max-width:600px\"><img style=\"max-width:122px\" src=\"https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/devtron/devtron-logo.png\" alt=\"ci-triggered\"></div></td></tr><tr><td colspan=\"3\"><div style=\"background-color:#e5f2ff;border-top-left-radius:8px;border-top-right-radius:8px;padding:20px 20px 16px 20px;display:flex;justify-content:space-between\"><div style=\"width:90%\"><div style=\"font-size:16px;line-height:24px;font-weight:600;margin-bottom:6px;color:#000a14\">Image promotion approval request</div><span style=\"font-size:14px;line-height:20px;color:#000a14\">{{eventTime}}</span><br><div><span style=\"font-size:14px;line-height:20px;color:#000a14\">by</span><span style=\"font-size:14px;line-height:20px;color:#06c;margin-left:4px\">{{triggeredBy}}</span></div></div><div><img src=\"https://cdn.devtron.ai/images/img_build_notification.png\" style=\"height:72px;width:72px\"></div></div></td></tr><tr><td colspan=\"3\"><div style=\"display:flex\"><div style=\"background-color:#e5f2ff;border-bottom-left-radius:8px;padding:0 0 20px 20px\"><a href=\"{{&artifactPromotionApprovalLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;background:#06c;color:#fff;border:1px solid transparent;cursor:pointer\">Approve</a></div><div style=\"width:90%;background-color:#e5f2ff;border-bottom-right-radius:8px;padding:0 0 20px 8px\"><a href=\"{{&artifactPromotionRequestViewLink}}\" style=\"height:32px;padding:7px 12px;line-height:32px;font-size:12px;font-weight:600;border-radius:4px;text-decoration:none;outline:0;min-width:64px;text-transform:capitalize;text-align:center;border:1px solid #d0d4d9;color:#06c;cursor:pointer;background-color:#fff\">View Request</a></div></div></td></tr><tr></tr><tr><td><br></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px\">Application</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{appName}}</div></td></tr><tr><td><br></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px\">Source</div></td><td colspan=\"2\"><div style=\"color:#3b444c;font-size:13px\">Target environment</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{promotionArtifactSource}}</div></td><td><div style=\"color:#000a14;font-size:14px\">{{envName}}</div></td></tr><tr><td colspan=\"3\"><div style=\"font-weight:600;margin-top:20px;width:100%;border-top:1px solid #edf1f5;padding:16px 0 12px;font-size:14px\">Image Details</div></td></tr><tr><td><div style=\"color:#3b444c;font-size:13px\">Image tag</div></td></tr><tr><td><div style=\"color:#000a14;font-size:14px\">{{imageTag}}</div></td></tr><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display:{{commentDisplayStyle}};\">Comment</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display:{{commentDisplayStyle}};\">{{comment}}</div></td></tr><tr><td><br></td></tr><tr><td colspan=\"3\"><div style=\"color:#3b444c;font-size:13px;display:{{tagDisplayStyle}};\">Tags</div></td></tr><tr><td colspan=\"3\"><div style=\"color:#000a14;font-size:14px;display:{};\">{{tags}}</div></td></tr><tr><td colspan=\"3\"><div style=\"border-top:1px solid #edf1f5;margin:20px 0 16px 0;height:1px\"></div></td></tr><tr><td colspan=\"2\" style=\"display:flex\"><span><a href=\"https://twitter.com/DevtronL\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/twitter_social_dark.png\"></div></a></span><span><a href=\"https://www.linkedin.com/company/devtron-labs/mycompany/\" target=\"_blank\" style=\"cursor:pointer;text-decoration:none;padding-right:12px;display:flex\"><div><img style=\"width:20px\" src=\"https://cdn.devtron.ai/images/linkedin_social_dark.png\"></div></a></span><span><a href=\"https://devtron.ai/blog/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline;padding-right:12px\">Blog</a></span><span><a href=\"https://devtron.ai/\" target=\"_blank\" style=\"color:#000a14;font-size:13px;line-height:20px;cursor:pointer;text-decoration:underline\">Website</a></span></td><td colspan=\"2\" style=\"text-align:right\"><div style=\"color:#767d84;font-size:13px;line-height:20px\">&copy; Devtron Labs 2024</div></td></tr></table>"
}');