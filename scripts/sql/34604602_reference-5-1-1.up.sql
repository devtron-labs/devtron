-- 1. Insert chart_ref if not exists
INSERT INTO "public"."chart_ref" (
    "location", "version", "deployment_strategy_path", 
    "is_default", "active", "created_on", "created_by", 
    "updated_on", "updated_by"
)
SELECT 
    'reference-chart_5-1-1', '5.1.1', 'pipeline-values.yaml', 
    'f', 't', now(), 1, now(), 1
WHERE NOT EXISTS (
    SELECT 1 
    FROM "public"."chart_ref" 
    WHERE "version" = '5.1.1' 
    AND "location" = 'reference-chart_5-1-1'
);

-- 2. Insert mappings based on the chart_ref above
INSERT INTO global_strategy_metadata_chart_ref_mapping (
    "global_strategy_metadata_id", "chart_ref_id", "active", 
    "created_on", "created_by", "updated_on", "updated_by", "default"
)
SELECT 
    m_ids.id,             
    cr.id,                
    true, now(), 1, now(), 1,
    (m_ids.id = 1)        
FROM 
    "public"."chart_ref" cr,
    (VALUES (1), (2), (3), (4)) AS m_ids(id) 
WHERE 
    cr.version = '5.1.1' AND cr.name IS NULL
    AND NOT EXISTS (
        SELECT 1 
        FROM global_strategy_metadata_chart_ref_mapping existing
        WHERE existing.global_strategy_metadata_id = m_ids.id
        AND existing.chart_ref_id = cr.id
    );