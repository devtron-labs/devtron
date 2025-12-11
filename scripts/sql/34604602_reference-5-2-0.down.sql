DELETE FROM global_strategy_metadata_chart_ref_mapping
WHERE chart_ref_id IN (
    SELECT id 
    FROM "public"."chart_ref" 
    WHERE "version" = '5.2.0' 
    AND "location" = 'reference-chart_5.2.0'
    AND "name" = 'Rollout Deployment'
)
AND global_strategy_metadata_id IN (1, 2, 3, 4);

-- 2. Remove the chart reference (Parent)
DELETE FROM "public"."chart_ref"
WHERE "version" = '5.2.0' 
AND "location" = 'reference-chart_5-2-0'
AND "name" = 'Rollout Deployment';
