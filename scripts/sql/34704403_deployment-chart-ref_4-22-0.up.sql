UPDATE chart_ref SET is_default=false;
INSERT INTO "public"."chart_ref" ("location", "version", "is_default", "active", "created_on", "created_by", "updated_on", "updated_by", "name", "deployment_strategy_path")
SELECT 'deployment-chart_4-22-0', '4.22.0', 't', 't', 'now()', 1, 'now()', 1, 'Deployment', 'pipeline-values.yaml'
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."chart_ref" WHERE "location" = 'deployment-chart_4-22-0'
);

INSERT INTO "public"."global_strategy_metadata_chart_ref_mapping" ("global_strategy_metadata_id", "chart_ref_id", "active", "default", "created_on", "created_by", "updated_on", "updated_by")
SELECT
    (SELECT "id" FROM "public"."global_strategy_metadata" WHERE "name" = 'ROLLING'),
    (SELECT "id" FROM "public"."chart_ref" WHERE "location" = 'deployment-chart_4-22-0'),
    true, true, 'now()', 1, 'now()', 1
WHERE NOT EXISTS (
    SELECT 1
    FROM "public"."global_strategy_metadata_chart_ref_mapping"
    WHERE "global_strategy_metadata_id" = (SELECT "id" FROM "public"."global_strategy_metadata" WHERE "name" = 'ROLLING')
      AND "chart_ref_id" = (SELECT "id" FROM "public"."chart_ref" WHERE "location" = 'deployment-chart_4-22-0')
);

INSERT INTO "public"."global_strategy_metadata_chart_ref_mapping" ("global_strategy_metadata_id", "chart_ref_id", "active", "default", "created_on", "created_by", "updated_on", "updated_by")
SELECT
    (SELECT "id" FROM "public"."global_strategy_metadata" WHERE "name" = 'RECREATE'),
    (SELECT "id" FROM "public"."chart_ref" WHERE "location" = 'deployment-chart_4-22-0'),
    true, false, 'now()', 1, 'now()', 1
WHERE NOT EXISTS (
    SELECT 1
    FROM "public"."global_strategy_metadata_chart_ref_mapping"
    WHERE "global_strategy_metadata_id" = (SELECT "id" FROM "public"."global_strategy_metadata" WHERE "name" = 'RECREATE')
      AND "chart_ref_id" = (SELECT "id" FROM "public"."chart_ref" WHERE "location" = 'deployment-chart_4-22-0')
);
