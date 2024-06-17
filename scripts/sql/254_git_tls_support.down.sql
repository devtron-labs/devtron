ALTER TABLE gitops_config DROP COLUMN IF EXISTS tls_cert,
ALTER TABLE gitops_config DROP COLUMN IF EXISTS tls_key ;
ALTER TABLE gitops_config DROP COLUMN IF EXISTS ca_cert;
ALTER TABLE gitops_config DROP COLUMN IF EXISTS enable_tls_verification;

ALTER TABLE git_provider DROP COLUMN IF EXISTS tls_cert,
ALTER TABLE git_provider DROP COLUMN IF EXISTS tls_key;
ALTER TABLE git_provider DROP COLUMN IF EXISTS ca_cert;
ALTER TABLE git_provider DROP COLUMN IF EXISTS enable_tls_verification;
