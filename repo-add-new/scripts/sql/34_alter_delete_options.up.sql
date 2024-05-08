ALTER TABLE chart_group ADD COLUMN deleted bool NOT NULL DEFAULT FALSE;

ALTER TABLE chart_repo ADD COLUMN deleted bool NOT NULL DEFAULT FALSE;

ALTER TABLE slack_config ADD COLUMN deleted bool NOT NULL DEFAULT FALSE;

ALTER TABLE ses_config ADD COLUMN deleted bool NOT NULL DEFAULT FALSE;

ALTER TABLE git_provider ADD COLUMN deleted bool NOT NULL DEFAULT FALSE;

ALTER TABLE team DROP CONSTRAINT team_name_key;

ALTER TABLE git_provider DROP CONSTRAINT git_provider_name_key;

ALTER TABLE git_provider DROP CONSTRAINT git_provider_url_key;

ALTER TABLE chart_group DROP CONSTRAINT chart_group_name_key;