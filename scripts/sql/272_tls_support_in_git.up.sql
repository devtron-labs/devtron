ALTER TABLE gitops_config ADD COLUMN tls_cert TEXT;
ALTER TABLE gitops_config ADD COLUMN tls_key TEXT;
ALTER TABLE gitops_config ADD COLUMN ca_cert TEXT;
ALTER TABLE gitops_config ADD COLUMN enable_tls_verification bool;

ALTER TABLE git_provider ADD COLUMN tls_key TEXT;
ALTER TABLE git_provider ADD COLUMN tls_cert TEXT;
ALTER TABLE git_provider ADD COLUMN ca_cert TEXT;
ALTER TABLE git_provider ADD COLUMN enable_tls_verification bool;
