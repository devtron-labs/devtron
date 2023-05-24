---- ALTER TABLE gitops_config - drop column
ALTER TABLE gitops_config
    DROP COLUMN IF EXISTS bitbucket_workspace_id,
    DROP COLUMN IF EXISTS bitbucket_project_key;