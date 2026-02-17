-- Rollback: Drop all Velero Backup & ClusterBackupRestore tables and sequences

-- Drop restore tables
DROP TABLE IF EXISTS cluster_backup_restore_status CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup_restore_status;

DROP TABLE IF EXISTS cluster_backup_restore CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup_restore;

-- Drop schedule tables
DROP TABLE IF EXISTS cluster_backup_schedule_status CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup_schedule_status;

DROP TABLE IF EXISTS cluster_backup_schedule CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup_schedule;

-- Drop backup tables
DROP TABLE IF EXISTS cluster_backup_status CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup_status;

DROP TABLE IF EXISTS cluster_backup CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup;

-- Drop location tables
DROP TABLE IF EXISTS cluster_backup_location_status CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup_location_status;

DROP TABLE IF EXISTS cluster_backup_location CASCADE;
DROP SEQUENCE IF EXISTS id_seq_cluster_backup_location;