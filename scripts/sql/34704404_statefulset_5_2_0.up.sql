-- 1. Idempotent Insert for chart_ref
INSERT INTO "public"."chart_ref" (
    "location", 
    "version",
    "deployment_strategy_path", 
    "is_default", 
    "active", 
    "created_on", 
    "created_by", 
    "updated_on", 
    "updated_by",
    "name"
)
SELECT 
    'statefulset-chart_5-2-0', 
    '5.2.0',
    'pipeline-values.yaml', 
    FALSE, 
    FALSE, 
    now(), 
    1, 
    now(), 
    1,
    'StatefulSet'
WHERE NOT EXISTS (
    SELECT 1 
    FROM "public"."chart_ref" 
    WHERE "location" = 'statefulset-chart_5-2-0'
);

-- 2. Idempotent Insert for Mapping (ROLLINGUPDATE)
INSERT INTO global_strategy_metadata_chart_ref_mapping (
    "global_strategy_metadata_id",
    "chart_ref_id", 
    "active",
    "default",
    "created_on", 
    "created_by", 
    "updated_on", 
    "updated_by"
)
SELECT 
    gsm.id, 
    cr.id, 
    TRUE, 
    TRUE, 
    now(), 
    1, 
    now(), 
    1
FROM 
    global_strategy_metadata gsm,
    chart_ref cr
WHERE 
    gsm.name = 'ROLLINGUPDATE' 
    AND cr.location = 'statefulset-chart_5-2-0'
    AND NOT EXISTS (
        SELECT 1 
        FROM global_strategy_metadata_chart_ref_mapping existing 
        WHERE existing.global_strategy_metadata_id = gsm.id 
          AND existing.chart_ref_id = cr.id
    );

-- 3. Idempotent Insert for Mapping (ONDELETE)
INSERT INTO global_strategy_metadata_chart_ref_mapping (
    "global_strategy_metadata_id",
    "chart_ref_id", 
    "active",
    "default",
    "created_on", 
    "created_by", 
    "updated_on", 
    "updated_by"
)
SELECT 
    gsm.id, 
    cr.id, 
    TRUE, 
    FALSE, 
    now(), 
    1, 
    now(), 
    1
FROM 
    global_strategy_metadata gsm,
    chart_ref cr
WHERE 
    gsm.name = 'ONDELETE' 
    AND cr.location = 'statefulset-chart_5-2-0'
    AND NOT EXISTS (
        SELECT 1 
        FROM global_strategy_metadata_chart_ref_mapping existing 
        WHERE existing.global_strategy_metadata_id = gsm.id 
          AND existing.chart_ref_id = cr.id
    );
