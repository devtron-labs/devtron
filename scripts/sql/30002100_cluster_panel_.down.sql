-- File: scripts/sql/30002100_cluster_panel_.down.sql

-- Drop Table
DROP TABLE IF EXISTS "public"."panel";

-- Drop Sequence
DROP SEQUENCE IF EXISTS id_seq_panel;