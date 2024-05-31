/*
 * Copyright (c) 2024. Devtron Inc.
 */

ALTER TABLE cluster add column insecure_skip_tls_verify boolean;
UPDATE cluster SET insecure_skip_tls_verify = true;