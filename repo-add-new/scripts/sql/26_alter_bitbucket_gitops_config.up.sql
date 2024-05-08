---- ALTER TABLE gitops_config - add column
ALTER TABLE gitops_config
    ADD COLUMN bitbucket_workspace_id TEXT,
    ADD COLUMN bitbucket_project_key TEXT;
