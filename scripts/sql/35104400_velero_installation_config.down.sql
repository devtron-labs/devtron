/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Drop indexes
DROP INDEX IF EXISTS "idx_velero_installation_config_status";
DROP INDEX IF EXISTS "idx_velero_installation_config_cluster_id";

-- Drop velero_installation_config table
DROP TABLE IF EXISTS "public"."velero_installation_config" CASCADE;

-- Remove vmware-tanzu chart repository (optional - commented out to preserve if used elsewhere)
-- DELETE FROM "public"."chart_repo" WHERE "name" = 'vmware-tanzu';
