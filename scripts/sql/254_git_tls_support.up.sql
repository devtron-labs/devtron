ALTER TABLE gitops_config ADD COLUMN tlsCert TEXT,
ALTER TABLE gitops_config ADD COLUMN tlsKey TEXT;
ALTER TABLE gitops_config ADD COLUMN caCert TEXT;

ALTER TABLE git_provider ADD COLUMN tlsCert TEXT,
ALTER TABLE git_provider ADD COLUMN tlsKey TEXT;
ALTER TABLE git_provider ADD COLUMN caCert TEXT;