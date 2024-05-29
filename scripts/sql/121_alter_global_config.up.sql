ALTER table global_cm_cs ADD secret_ingestion_for VARCHAR(50);
/*
 * Copyright (c) 2024. Devtron Inc.
 */

--setting type as CI/CD because secrets until this release should be available in both CI and CD pipelines
UPDATE global_cm_cs SET secret_ingestion_for='CI/CD';