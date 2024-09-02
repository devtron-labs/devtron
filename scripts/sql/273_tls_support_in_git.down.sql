ALTER TABLE "git_provider"
    DROP COLUMN tls_cert ,
    DROP COLUMN tls_key ,
    DROP COLUMN ca_cert,
    DROP COLUMN enable_tls_verification ;

ALTER TABLE "gitops_config"
DROP COLUMN tls_cert ,
    DROP COLUMN tls_key ,
    DROP COLUMN ca_cert,
    DROP COLUMN enable_tls_verification ;