ALTER TABLE global_cm_cs ADD pipeline_type VARCHAR(50);
--setting type as volume because until this release only volume type was supported
UPDATE global_cm_cs SET pipeline_type='CI/CD';