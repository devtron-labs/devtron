--- Drop
ALTER TABLE "public"."user_deployment_request"
    DROP COLUMN IF EXISTS "helm_redeployment_request",
    DROP COLUMN IF EXISTS "helm_take_release_ownership";