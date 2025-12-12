-- Velero Backup & ClusterBackupRestore Integration - Final Schema
-- This migration creates all tables for Velero backup and restore functionality
-- with merged location tables and JSONB configuration/status fields

-- =====================================================
-- Backup Location Tables (Unified)
-- =====================================================

-- Unified backup location table (storage + snapshot locations)
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup_location;

CREATE TABLE IF NOT EXISTS cluster_backup_location (
    id                      INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup_location'::regclass),
    name                    VARCHAR(255) NOT NULL,
    location_type           VARCHAR(20) NOT NULL, -- 'storage' or 'snapshot'
    provider                VARCHAR(100) NOT NULL,
    
    -- Storage location specific fields (NULL for snapshot locations)
    bucket                  VARCHAR(255),
    prefix                  VARCHAR(255),
    is_default              BOOLEAN DEFAULT false,
    
    -- Common fields
    config                  JSONB,
    ca_cert                 BYTEA,
    access_mode             VARCHAR(20), -- 'readwrite', 'readonly'
    backup_sync_period      VARCHAR(255), -- e.g., "30s"
    validation_frequency    VARCHAR(255), -- e.g., "10m"

    -- Credential fields
    credential_type         VARCHAR(20), -- 'existing_secret', 'create_secret'
    credential_secret_name  VARCHAR(255),
    credential_secret_key   VARCHAR(255),
    credential_secret_data  TEXT,

    active                  BOOLEAN NOT NULL DEFAULT true,
    created_on              TIMESTAMPTZ NOT NULL,
    created_by              INTEGER NOT NULL,
    updated_on              TIMESTAMPTZ NOT NULL,
    updated_by              INTEGER NOT NULL,
    
    CONSTRAINT pk_cluster_backup_location PRIMARY KEY (id),
    CONSTRAINT chk_cluster_backup_location_type CHECK (location_type IN ('storage', 'snapshot'))
);

-- Unique index on name and location_type
CREATE UNIQUE INDEX idx_cluster_backup_location_name_type ON cluster_backup_location(name, location_type) WHERE active = true;
-- Indexes for faster querying
CREATE INDEX idx_cluster_backup_location_type ON cluster_backup_location(location_type);
CREATE INDEX idx_cluster_backup_location_active ON cluster_backup_location(active);

-- Unified location status table
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup_location_status;

CREATE TABLE IF NOT EXISTS cluster_backup_location_status (
    id                          INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup_location_status'::regclass),
    cluster_backup_location_id  INTEGER NOT NULL,
    cluster_id                  INTEGER NOT NULL,
    location_type               VARCHAR(20) NOT NULL, -- 'storage' or 'snapshot'
    phase                       VARCHAR(50),
    last_synced_time            TIMESTAMPTZ,
    last_validation_time        TIMESTAMPTZ,
    message                     TEXT,
    created_on                  TIMESTAMPTZ NOT NULL,
    updated_on                  TIMESTAMPTZ NOT NULL,
    
    CONSTRAINT pk_cluster_backup_location_status PRIMARY KEY (id),
    CONSTRAINT fk_cluster_backup_location_status_location FOREIGN KEY (cluster_backup_location_id) REFERENCES cluster_backup_location(id) ON DELETE CASCADE,
    CONSTRAINT fk_cluster_backup_location_status_cluster FOREIGN KEY (cluster_id) REFERENCES cluster(id),
    CONSTRAINT uq_cluster_backup_location_status UNIQUE (cluster_backup_location_id, cluster_id)
);

CREATE INDEX idx_cluster_backup_location_status_location_id ON cluster_backup_location_status(cluster_backup_location_id);
CREATE INDEX idx_cluster_backup_location_status_cluster_id ON cluster_backup_location_status(cluster_id);

-- =====================================================
-- Backup Tables
-- =====================================================

-- Backup table with JSONB configuration
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup;

CREATE TABLE IF NOT EXISTS cluster_backup (
    id                          INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup'::regclass),
    cluster_id                  INTEGER NOT NULL,
    name                        VARCHAR(255) NOT NULL,
    description                 VARCHAR(300),
    storage_location_id         INTEGER NOT NULL,
    snapshot_location_ids       INTEGER[],
    cluster_backup_schedule_id  INTEGER,
    configuration               JSONB NOT NULL,
    configuration_version       INTEGER NOT NULL DEFAULT 1,
    active                      BOOLEAN NOT NULL DEFAULT true,
    created_on                  TIMESTAMPTZ NOT NULL,
    created_by                  INTEGER NOT NULL,
    updated_on                  TIMESTAMPTZ NOT NULL,
    updated_by                  INTEGER NOT NULL,
    
    CONSTRAINT pk_cluster_backup PRIMARY KEY (id),
    CONSTRAINT fk_cluster_backup_cluster FOREIGN KEY (cluster_id) REFERENCES cluster(id),
    CONSTRAINT fk_cluster_backup_storage_location FOREIGN KEY (storage_location_id) REFERENCES cluster_backup_location(id)
);

-- Unique index on name and cluster_id
CREATE UNIQUE INDEX idx_cluster_backup_name_cluster ON cluster_backup(name, cluster_id) WHERE active = true;
-- Indexes for faster querying
CREATE INDEX idx_cluster_backup_cluster_id ON cluster_backup(cluster_id);
CREATE INDEX idx_cluster_backup_storage_location_id ON cluster_backup(storage_location_id);
CREATE INDEX idx_cluster_backup_schedule_id ON cluster_backup(cluster_backup_schedule_id);
CREATE INDEX idx_cluster_backup_active ON cluster_backup(active);
CREATE INDEX idx_cluster_backup_configuration ON cluster_backup USING gin(configuration);

-- Backup status table with JSONB status_detail
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup_status;

CREATE TABLE IF NOT EXISTS cluster_backup_status (
    id                      INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup_status'::regclass),
    cluster_backup_id       INTEGER NOT NULL,
    phase                   VARCHAR(50),
    status_detail           JSONB,
    status_detail_version   INTEGER NOT NULL DEFAULT 1,
    created_on              TIMESTAMPTZ NOT NULL,
    updated_on              TIMESTAMPTZ NOT NULL,
    started_on              TIMESTAMPTZ,
    finished_on             TIMESTAMPTZ,
    expires_at              TIMESTAMPTZ,
    
    CONSTRAINT pk_cluster_backup_status PRIMARY KEY (id),
    CONSTRAINT fk_cluster_backup_status_backup FOREIGN KEY (cluster_backup_id) REFERENCES cluster_backup(id) ON DELETE CASCADE,
    CONSTRAINT uq_cluster_backup_status UNIQUE (cluster_backup_id)
);

CREATE INDEX idx_cluster_backup_status_cluster_backup_id ON cluster_backup_status(cluster_backup_id);
CREATE INDEX idx_cluster_backup_status_detail ON cluster_backup_status USING gin(status_detail);

-- =====================================================
-- Schedule Tables
-- =====================================================

-- Schedule table with JSONB configuration
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup_schedule;

CREATE TABLE IF NOT EXISTS cluster_backup_schedule (
    id                              INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup_schedule'::regclass),
    cluster_id                      INTEGER NOT NULL,
    name                            VARCHAR(255) NOT NULL,
    description                     VARCHAR(300),
    schedule                        VARCHAR(100) NOT NULL, -- Cron expression
    storage_location_id             INTEGER NOT NULL,
    snapshot_location_ids           INTEGER[],
    configuration                   JSONB NOT NULL,
    configuration_version           INTEGER NOT NULL DEFAULT 1,
    paused                          BOOLEAN NOT NULL DEFAULT false,
    use_owner_references_in_backup  BOOLEAN,
    skip_immediately                BOOLEAN,
    active                          BOOLEAN NOT NULL DEFAULT true,
    created_on                      TIMESTAMPTZ NOT NULL,
    created_by                      INTEGER NOT NULL,
    updated_on                      TIMESTAMPTZ NOT NULL,
    updated_by                      INTEGER NOT NULL,
    
    CONSTRAINT pk_cluster_backup_schedule PRIMARY KEY (id),
    CONSTRAINT fk_cluster_backup_schedule_cluster FOREIGN KEY (cluster_id) REFERENCES cluster(id),
    CONSTRAINT fk_cluster_backup_schedule_storage_location FOREIGN KEY (storage_location_id) REFERENCES cluster_backup_location(id)
);

-- Unique index on name and cluster_id
CREATE UNIQUE INDEX idx_cluster_backup_schedule_name_cluster ON cluster_backup_schedule(name, cluster_id) WHERE active = true;
-- Indexes for faster querying
CREATE INDEX idx_cluster_backup_schedule_cluster_id ON cluster_backup_schedule(cluster_id);
CREATE INDEX idx_cluster_backup_schedule_storage_location_id ON cluster_backup_schedule(storage_location_id);
CREATE INDEX idx_cluster_backup_schedule_active ON cluster_backup_schedule(active);
CREATE INDEX idx_cluster_backup_schedule_configuration ON cluster_backup_schedule USING gin(configuration);

-- Add foreign key from backup to schedule
ALTER TABLE cluster_backup
    ADD CONSTRAINT fk_cluster_backup_schedule
    FOREIGN KEY (cluster_backup_schedule_id)
    REFERENCES cluster_backup_schedule(id);

-- Schedule status table with JSONB status_detail
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup_schedule_status;

CREATE TABLE IF NOT EXISTS cluster_backup_schedule_status (
    id                          INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup_schedule_status'::regclass),
    cluster_backup_schedule_id  INTEGER NOT NULL,
    phase                       VARCHAR(50),
    last_backup_time            TIMESTAMPTZ,
    status_detail               JSONB,
    status_detail_version       INTEGER NOT NULL DEFAULT 1,
    created_on                  TIMESTAMPTZ NOT NULL,
    updated_on                  TIMESTAMPTZ NOT NULL,
    
    CONSTRAINT pk_cluster_backup_schedule_status PRIMARY KEY (id),
    CONSTRAINT fk_cluster_backup_schedule_status_schedule FOREIGN KEY (cluster_backup_schedule_id) REFERENCES cluster_backup_schedule(id) ON DELETE CASCADE,
    CONSTRAINT uq_cluster_backup_schedule_status UNIQUE (cluster_backup_schedule_id)
);

CREATE INDEX idx_cluster_backup_schedule_status_schedule_id ON cluster_backup_schedule_status(cluster_backup_schedule_id);
CREATE INDEX idx_cluster_backup_schedule_status_detail ON cluster_backup_schedule_status USING gin(status_detail);

-- =====================================================
-- ClusterBackupRestore Tables
-- =====================================================

-- ClusterBackupRestore table with JSONB configuration
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup_restore;

CREATE TABLE IF NOT EXISTS cluster_backup_restore (
    id                          INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup_restore'::regclass),
    cluster_id                  INTEGER NOT NULL,
    name                        VARCHAR(255) NOT NULL,
    description                 VARCHAR(300),
    cluster_backup_id           INTEGER,
    cluster_backup_schedule_id  INTEGER,
    configuration               JSONB NOT NULL,
    configuration_version       INTEGER NOT NULL DEFAULT 1,
    active                      BOOLEAN NOT NULL DEFAULT true,
    created_on                  TIMESTAMPTZ NOT NULL,
    created_by                  INTEGER NOT NULL,
    updated_on                  TIMESTAMPTZ NOT NULL,
    updated_by                  INTEGER NOT NULL,

    CONSTRAINT pk_cluster_backup_restore PRIMARY KEY (id),
    CONSTRAINT fk_cluster_backup_restore_cluster FOREIGN KEY (cluster_id) REFERENCES cluster(id),
    CONSTRAINT fk_cluster_backup_restore_backup FOREIGN KEY (cluster_backup_id) REFERENCES cluster_backup(id),
    CONSTRAINT fk_cluster_backup_restore_schedule FOREIGN KEY (cluster_backup_schedule_id) REFERENCES cluster_backup_schedule(id)
);

-- Unique index on name and cluster_id
CREATE UNIQUE INDEX idx_cluster_backup_restore_name_cluster ON cluster_backup_restore(name, cluster_id) WHERE active = true;
-- Indexes for faster querying
CREATE INDEX idx_cluster_backup_restore_cluster_id ON cluster_backup_restore(cluster_id);
CREATE INDEX idx_cluster_backup_restore_cluster_backup_id ON cluster_backup_restore(cluster_backup_id);
CREATE INDEX idx_cluster_backup_restore_cluster_backup_schedule_id ON cluster_backup_restore(cluster_backup_schedule_id);
CREATE INDEX idx_cluster_backup_restore_active ON cluster_backup_restore(active);
CREATE INDEX idx_cluster_backup_restore_configuration ON cluster_backup_restore USING gin(configuration);

-- ClusterBackupRestore status table with JSONB status_detail
CREATE SEQUENCE IF NOT EXISTS id_seq_cluster_backup_restore_status;

CREATE TABLE IF NOT EXISTS cluster_backup_restore_status (
    id                          INTEGER NOT NULL DEFAULT nextval('id_seq_cluster_backup_restore_status'::regclass),
    cluster_backup_restore_id   INTEGER NOT NULL,
    phase                       VARCHAR(50),
    status_detail               JSONB,
    status_detail_version       INTEGER NOT NULL DEFAULT 1,
    created_on                  TIMESTAMPTZ NOT NULL,
    updated_on                  TIMESTAMPTZ NOT NULL,
    started_on                  TIMESTAMPTZ,
    finished_on                 TIMESTAMPTZ,
    
    CONSTRAINT pk_cluster_backup_restore_status PRIMARY KEY (id),
    CONSTRAINT fk_cluster_backup_restore_status_restore FOREIGN KEY (cluster_backup_restore_id) REFERENCES cluster_backup_restore(id) ON DELETE CASCADE,
    CONSTRAINT uq_cluster_backup_restore_status UNIQUE (cluster_backup_restore_id)
);

CREATE INDEX idx_cluster_backup_restore_status_restore_id ON cluster_backup_restore_status(cluster_backup_restore_id);
CREATE INDEX idx_cluster_backup_restore_status_detail ON cluster_backup_restore_status USING gin(status_detail);