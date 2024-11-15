-- 29902101_add_is_prod_column_to_cluster_table.down.sql
ALTER TABLE cluster DROP COLUMN IF EXISTS is_prod;