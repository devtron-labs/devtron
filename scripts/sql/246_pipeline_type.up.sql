/*
 * Copyright (c) 2024. Devtron Inc.
 */

UPDATE   "public"."ci_pipeline"  SET ci_pipeline_type='LINKED' WHERE ci_pipeline_type='CI_EXTERNAL';

UPDATE "public"."ci_pipeline"
SET ci_pipeline_type = 'CI_BUILD'
    FROM app
WHERE ci_pipeline.app_id = app.id AND ci_pipeline.ci_pipeline_type IS NULL AND app.app_type in( 0,2);