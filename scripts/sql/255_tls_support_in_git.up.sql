ALTER TABLE "git_provider"
    ADD COLUMN tls_cert text,
    ADD COLUMN tls_key text,
    ADD COLUMN ca_cert text,
    ADD COLUMN enable_tls_verification bool;

ALTER TABLE "gitops_config"
    ADD COLUMN tls_cert text,
    ADD COLUMN tls_key text,
    ADD COLUMN ca_cert text,
    ADD COLUMN enable_tls_verification bool;

