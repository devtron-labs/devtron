-- First, safely insert the chart reference if it doesn't already exist.
INSERT INTO "public"."chart_ref" ("location", "version", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "name", "deployment_strategy_path")
SELECT 'cronjob-chart_1-6-0', '1.6.0', 'f', 't', 'now()', 1, 'now()', 1, 'Job & CronJob', 'pipeline-values.yaml'
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."chart_ref" WHERE "location" = 'cronjob-chart_1-6-0'
);

-- Next, safely insert the mapping if it doesn't already exist.
INSERT INTO "public"."global_strategy_metadata_chart_ref_mapping" ("global_strategy_metadata_id", "chart_ref_id", "active", "default", "created_on", "created_by", "updated_on", "updated_by")
SELECT
    (SELECT "id" FROM "public"."global_strategy_metadata" WHERE "name" = 'ROLLING'),
    (SELECT "id" FROM "public"."chart_ref" WHERE "location" = 'cronjob-chart_1-6-0'),
    true, true, 'now()', 1, 'now()', 1
WHERE NOT EXISTS (
    SELECT 1
    FROM "public"."global_strategy_metadata_chart_ref_mapping"
    WHERE "global_strategy_metadata_id" = (SELECT "id" FROM "public"."global_strategy_metadata" WHERE "name" = 'ROLLING')
      AND "chart_ref_id" = (SELECT "id" FROM "public"."chart_ref" WHERE "location" = 'cronjob-chart_1-6-0')
);
