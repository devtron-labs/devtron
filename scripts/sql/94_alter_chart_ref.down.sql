ALTER TABLE "chart_ref" DROP COLUMN "file_path_containing_strategy";
ALTER TABLE "chart_ref" DROP COLUMN "json_path_for_strategy";
ALTER TABLE "chart_ref" DROP COLUMN "is_app_metrics_supported";

DROP TABLE IF EXISTS "global_strategy_metadata" CASCADE;

DROP SEQUENCE IF EXISTS "id_seq_global_strategy_metadata";

DROP TABLE IF EXISTS "global_strategy_metadata_chart_ref_mapping" CASCADE;

DROP SEQUENCE IF EXISTS "id_seq_global_strategy_metadata_chart_ref_mapping";
