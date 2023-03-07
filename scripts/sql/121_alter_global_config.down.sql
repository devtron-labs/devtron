ALTER TABLE global_cm_cs ADD pipeline_type VARCHAR(50);
--setting type as CI/CD because secrets until this release should be available in both CI and CD pipelines
UPDATE global_cm_cs SET pipeline_type='CI/CD';