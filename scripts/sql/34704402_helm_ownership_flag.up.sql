--- Adding helm_take_ownership column to user_deployment_request table
ALTER TABLE "public"."user_deployment_request"
    ADD COLUMN IF NOT EXISTS "helm_redeployment_request" boolean DEFAULT false,
    ADD COLUMN IF NOT EXISTS "helm_take_release_ownership" boolean DEFAULT false;