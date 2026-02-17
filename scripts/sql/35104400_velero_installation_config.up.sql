/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Create velero_installation_config table for tracking Velero installation status
CREATE TABLE IF NOT EXISTS "public"."velero_installation_config" (
    "id" SERIAL PRIMARY KEY,
    "cluster_id" int4 NOT NULL,
    "namespace" varchar(255) NOT NULL DEFAULT 'velero',
    "status" varchar(50) NOT NULL,
    "error_message" text,
    "created_on" timestamptz NOT NULL DEFAULT NOW(),
    "created_by" int4 NOT NULL,
    "updated_on" timestamptz NOT NULL DEFAULT NOW(),
    "updated_by" int4 NOT NULL,
    CONSTRAINT "fk_velero_installation_config_cluster_id" FOREIGN KEY ("cluster_id") REFERENCES "public"."cluster"("id"),
    CONSTRAINT "uq_velero_installation_config_cluster_id" UNIQUE ("cluster_id")
);

-- Add index for faster lookups
CREATE INDEX IF NOT EXISTS "idx_velero_installation_config_cluster_id" ON "public"."velero_installation_config" ("cluster_id");
CREATE INDEX IF NOT EXISTS "idx_velero_installation_config_status" ON "public"."velero_installation_config" ("status");

-- Add chart repository for Velero if it doesn't exist
INSERT INTO "public"."chart_repo" ("name", "url", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "external", "auth_mode", "deleted", "allow_insecure_connection")
SELECT 'vmware-tanzu', 'https://vmware-tanzu.github.io/helm-charts', 'f', 't', NOW(), 1, NOW(), 1, 't', 'ANONYMOUS', 'f', 't'
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."chart_repo" 
    WHERE "name" = 'vmware-tanzu' AND "active" = true AND "deleted" = false
);
