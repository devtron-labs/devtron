/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Create new cost_module_enabled column in cluster table
ALTER TABLE "public"."cluster" ADD COLUMN IF NOT EXISTS "cost_module_enabled" bool NOT NULL DEFAULT false;

-- Create sequence for cost_custom_view
CREATE SEQUENCE IF NOT EXISTS id_seq_cost_custom_view;

-- Create cost_custom_view table
CREATE TABLE IF NOT EXISTS "public"."cost_custom_view" (
    "id"          int4 NOT NULL DEFAULT nextval('id_seq_cost_custom_view'::regclass),
    "name"        varchar(255) NOT NULL,
    "identifier"  varchar(255) NOT NULL,
    "description" text,
    "filters"     text NOT NULL,
    "active"   bool NOT NULL DEFAULT true,
    "cluster_id"  int4 NOT NULL,
    "created_on"  timestamptz DEFAULT NOW(),
    "created_by"  int4,
    "updated_on"  timestamptz,
    "updated_by"  int4,
    PRIMARY KEY ("id")
);

-- Indexes for cost_custom_view
CREATE INDEX IF NOT EXISTS idx_cost_custom_view_name ON "public"."cost_custom_view"("name") WHERE "active" = true;

-- Unique constraints to ensure name uniqueness within cluster and identifier uniqueness globally
CREATE UNIQUE INDEX IF NOT EXISTS idx_cost_custom_view_name_cluster_unique
    ON "public"."cost_custom_view"("name", "cluster_id")
    WHERE "active" = true;

CREATE UNIQUE INDEX IF NOT EXISTS idx_cost_custom_view_identifier_unique
    ON "public"."cost_custom_view"("identifier")
    WHERE "active" = true;

-- Add comments for documentation
COMMENT ON TABLE "public"."cost_custom_view" IS 'User-defined custom views for cost analysis and filtering';
COMMENT ON COLUMN "public"."cost_custom_view"."filters" IS 'JSON string containing filter criteria for the custom view';

INSERT INTO "public"."chart_repo" ( "name", "url", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "external","auth_mode","deleted","allow_insecure_connection")
SELECT 'opencost', 'https://opencost.github.io/opencost-helm-chart', 'f', 't', now(), 1, now(), 1, 't','ANONYMOUS','f','t'
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."chart_repo" WHERE "name" = 'opencost' and "active"=true and "deleted"=false
);