alter table pipeline DROP column IF EXISTS user_approval_config;

update public.deployment_approval_request set active = false;

alter table cd_workflow_runner drop column IF EXISTS deployment_approval_request_id;

DROP INDEX "public"."role_unique";

Alter table roles DROP column IF EXISTS approver;

CREATE UNIQUE INDEX IF NOT EXISTS "role_unique" ON "public"."roles" USING BTREE ("role","access_type");