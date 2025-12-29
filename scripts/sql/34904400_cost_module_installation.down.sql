/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Remove cost module installation fields from cluster table
ALTER TABLE "public"."cluster" DROP COLUMN IF EXISTS "cost_module_installation_status";
ALTER TABLE "public"."cluster" DROP COLUMN IF EXISTS "cost_module_installation_error";
ALTER TABLE "public"."cluster" DROP COLUMN IF EXISTS "cost_config";


-- Drop default currency setting
DELETE FROM "public"."attributes" WHERE key = 'defaultCurrency';