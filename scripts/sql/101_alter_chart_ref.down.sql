DELETE FROM "public"."chart_ref" WHERE ("location" = 'deployment-chart_1-0-0' AND "version" = '1.0.0');

ALTER TABLE "chart_ref" DROP COLUMN "deployment_strategy_path";
ALTER TABLE "chart_ref" DROP COLUMN "json_path_for_strategy";
ALTER TABLE "chart_ref" DROP COLUMN "is_app_metrics_supported";

DROP TABLE IF EXISTS "global_strategy_metadata" CASCADE;

DROP SEQUENCE IF EXISTS "id_seq_global_strategy_metadata";

DROP TABLE IF EXISTS "global_strategy_metadata_chart_ref_mapping" CASCADE;

DROP SEQUENCE IF EXISTS "id_seq_global_strategy_metadata_chart_ref_mapping";
