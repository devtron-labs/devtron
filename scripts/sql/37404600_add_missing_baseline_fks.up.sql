-- Two FK constraints that are part of devtron OSS's baseline schema
-- (devtron/scripts/sql/1_insert.up.sql) but were found missing on at least one real
-- installation during db-housekeeping testing (devtron-services-enterprise PR #443).
-- db-housekeeping's database/schemaassert.go refuses to run (AssertSchema) when either
-- is absent, since its cascade-delete logic depends on this exact FK graph.
--
-- Wrapped in existence checks so this migration is a no-op on installations that already
-- have these constraints (the common case) and only takes effect where they're missing.
-- Both were verified safe to add on the affected cluster: zero rows violated either FK
-- before adding it.

DO $$
BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'pipeline_config_override_cd_workflow_id_fkey'
  ) THEN
    ALTER TABLE pipeline_config_override
      ADD CONSTRAINT pipeline_config_override_cd_workflow_id_fkey
      FOREIGN KEY (cd_workflow_id) REFERENCES cd_workflow(id);
  END IF;

  IF NOT EXISTS (
    SELECT 1 FROM pg_constraint WHERE conname = 'ci_artifact_parent_ci_artifact_fkey'
  ) THEN
    ALTER TABLE ci_artifact
      ADD CONSTRAINT ci_artifact_parent_ci_artifact_fkey
      FOREIGN KEY (parent_ci_artifact) REFERENCES ci_artifact(id);
  END IF;
END $$;
