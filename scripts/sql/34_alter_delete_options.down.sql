ALTER TABLE chart_group DROP COLUMN IF EXISTS deleted;

ALTER TABLE chart_repo DROP COLUMN IF EXISTS deleted;

ALTER TABLE slack_config DROP COLUMN IF EXISTS deleted;

ALTER TABLE ses_config DROP COLUMN IF EXISTS deleted;

ALTER TABLE git_provider DROP COLUMN IF EXISTS deleted;

ALTER TABLE team ADD CONSTRAINT team_name_key UNIQUE (name);

ALTER TABLE git_provider ADD CONSTRAINT git_provider_name_key UNIQUE (name);

ALTER TABLE git_provider ADD CONSTRAINT git_provider_url_key UNIQUE (url);

ALTER TABLE chart_group ADD CONSTRAINT chart_group_name_key UNIQUE (name);