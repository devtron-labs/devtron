/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Add cost module installation status and configuration fields to cluster table
ALTER TABLE "public"."cluster" ADD COLUMN IF NOT EXISTS "cost_module_installation_status" varchar(50);
ALTER TABLE "public"."cluster" ADD COLUMN IF NOT EXISTS "cost_module_installation_error" text;

ALTER TABLE "public"."cluster" ADD COLUMN IF NOT EXISTS "gpu_installation_status" varchar(50);
ALTER TABLE "public"."cluster" ADD COLUMN IF NOT EXISTS "gpu_installation_error" text;

ALTER TABLE "public"."cluster" ADD COLUMN IF NOT EXISTS "cost_config" jsonb;

-- Add default currency setting (only if it doesn't already exist)
INSERT INTO "public"."attributes"(key, value, active, created_on, created_by, updated_on, updated_by)
SELECT 'defaultCurrency', 'USD', 't', NOW(), 1, NOW(), 1
    WHERE NOT EXISTS (
    SELECT 1 FROM "public"."attributes"
    WHERE key = 'defaultCurrency'
);

INSERT INTO "public"."chart_repo" ( "name", "url", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "external","auth_mode","deleted","allow_insecure_connection")
SELECT 'gpu-helm-charts', 'https://nvidia.github.io/dcgm-exporter/helm-charts', 'f', 't', now(), 1, now(), 1, 't','ANONYMOUS','f','t'
    WHERE NOT EXISTS (
    SELECT 1 FROM "public"."chart_repo" WHERE "name" = 'gpu-helm-charts' and "active"=true and "deleted"=false
);