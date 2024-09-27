---
--- Add is_artifact_uploaded flag to ci_workflow and cd_workflow_runner tables
---

ALTER TABLE "public"."ci_workflow"
    ADD COLUMN IF NOT EXISTS is_artifact_uploaded BOOLEAN;
ALTER TABLE "public"."cd_workflow_runner"
    ADD COLUMN IF NOT EXISTS is_artifact_uploaded BOOLEAN,
    ADD COLUMN IF NOT EXISTS cd_artifact_location varchar(256);

---
--- Drop ci_workflow_config table; redundant table
---

DROP TABLE IF EXISTS "public"."ci_workflow_config";

---
--- Drop cd_workflow_config table; redundant table
---

DROP TABLE IF EXISTS "public"."cd_workflow_config";
