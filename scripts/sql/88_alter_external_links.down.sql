ALTER TABLE "public"."external_link" DROP COLUMN is_editable;
ALTER TABLE "public"."external_link" DROP COLUMN description;

ALTER TABLE "public"."external_link_monitoring_tool" DROP COLUMN category;

DROP TABLE "public"."external_link_identifier_mapping";