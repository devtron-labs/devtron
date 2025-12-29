/*
 * Copyright (c) 2024. Devtron Inc.
 */

-- Drop indexes and constraints first
DROP INDEX IF EXISTS "public"."idx_cost_custom_view_name";
DROP INDEX IF EXISTS "public"."idx_cost_custom_view_name_cluster_unique";
DROP INDEX IF EXISTS "public"."idx_cost_custom_view_identifier_unique";

-- Drop table
DROP TABLE IF EXISTS "public"."cost_custom_view" CASCADE;

-- Drop sequences
DROP SEQUENCE IF EXISTS id_seq_cost_custom_view;

-- Drop column from cluster table
ALTER TABLE "public"."cluster" DROP COLUMN IF EXISTS "cost_module_enabled";
