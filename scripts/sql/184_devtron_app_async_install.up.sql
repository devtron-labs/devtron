-- cd_workflow_runner.message has a limit of 256 characters. migrating to
ALTER TABLE cd_workflow_runner
ALTER COLUMN message TYPE VARCHAR(1000);