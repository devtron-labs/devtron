ALTER TABLE "public"."external_link" DROP COLUMN is_editable;
ALTER TABLE "public"."external_link" DROP COLUMN description;

ALTER TABLE "public"."external_link_monitoring_tool" DROP COLUMN category;

ALTER TABLE IF EXISTS "public"."external_link_identifier_mapping" DROP COLUMN "type";
ALTER TABLE IF EXISTS "public"."external_link_identifier_mapping" DROP COLUMN "identifier";
ALTER TABLE IF EXISTS "public"."external_link_identifier_mapping" DROP COLUMN "env_id";
ALTER TABLE IF EXISTS "public"."external_link_identifier_mapping" DROP COLUMN "app_id";

ALTER SEQUENCE IF EXISTS id_seq_external_link_identifier_mapping RENAME TO id_seq_external_link_cluster_mapping;
ALTER TABLE IF EXISTS "public"."external_link_identifier_mapping" RENAME TO external_link_cluster_mapping;
