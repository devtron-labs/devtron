--ADD Columns is_editable and description in external_link table
ALTER TABLE "public"."external_link" ADD COLUMN is_editable bool NOT NULL DEFAULT false;
ALTER TABLE "public"."external_link" ADD COLUMN description text;

--ADD column category for external_link_monitoring_tool
ALTER TABLE "public"."external_link_monitoring_tool" ADD COLUMN category int4;
ALTER TABLE IF EXISTS "public"."external_link_cluster_mapping" ADD COLUMN "type" int4 NOT NULL DEFAULT 0;
ALTER TABLE IF EXISTS "public"."external_link_cluster_mapping" ADD COLUMN "identifier" varchar(255) NOT NULL DEFAULT '';
ALTER TABLE IF EXISTS "public"."external_link_cluster_mapping" ADD COLUMN "env_id" int4 NOT NULL DEFAULT 0;
ALTER TABLE IF EXISTS "public"."external_link_cluster_mapping" ADD COLUMN "app_id" int4 NOT NULL DEFAULT 0;

ALTER SEQUENCE IF EXISTS id_seq_external_link_cluster_mapping RENAME TO id_seq_external_link_identifier_mapping;
ALTER TABLE IF EXISTS "public"."external_link_cluster_mapping" RENAME TO external_link_identifier_mapping;




