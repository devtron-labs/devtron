BEGIN;

CREATE SEQUENCE IF NOT EXISTS krr_scan_request_id_seq;

CREATE TABLE IF NOT EXISTS krr_scan_request
(
    "id"              INTEGER PRIMARY KEY DEFAULT nextval('krr_scan_request_id_seq'),
    "kind"            VARCHAR(250) NOT NULL,
    "api_version"     VARCHAR(250) NOT NULL,
    "name"            VARCHAR(250) NOT NULL,
    "namespace"       VARCHAR(250) NOT NULL,
    "cluster_id"      INTEGER      NOT NULL,
    "created_on"      TIMESTAMPTZ,
    "created_by"      INTEGER,
    "updated_on"      TIMESTAMPTZ,
    "updated_by"      INTEGER,
    CONSTRAINT cluster_id_krr_scan_request_fk FOREIGN KEY (cluster_id) REFERENCES cluster (id)
);


CREATE SEQUENCE IF NOT EXISTS krr_scan_history_id_seq;

CREATE TABLE IF NOT EXISTS krr_scan_history
(
    "id"                        INTEGER PRIMARY KEY DEFAULT nextval('krr_scan_history_id_seq'),
    "krr_scan_request"          INTEGER       NOT NULL,
    "container"                 VARCHAR(1000) NOT NULL,
    "scanned_on"                TIMESTAMPTZ   NOT NULL,
    "scanned_by"                INTEGER,
    "resource_type"             VARCHAR(50) CHECK (resource_type IN ('CPU', 'MEMORY')),
    "recommended_strategy_name" VARCHAR(250),
    "recommended_request"       DOUBLE PRECISION,
    "recommended_limits"        DOUBLE PRECISION,
    "unit"                      VARCHAR(50),
    "evaluation"                JSONB,
    "recommendation"            JSONB,
    "created_on"                TIMESTAMPTZ,
    "created_by"                INTEGER,
    "updated_on"                TIMESTAMPTZ,
    "updated_by"                INTEGER,
    CONSTRAINT krr_scan_request_fk FOREIGN KEY (krr_scan_request) REFERENCES krr_scan_request (id)
);

END;