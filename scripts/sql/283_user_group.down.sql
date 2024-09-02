DROP INDEX  IF EXISTS idx_unique_user_group_user_id;
DROP TABLE IF EXISTS "public"."user_group_mapping";
DROP INDEX IF EXISTS idx_unique_user_group_name;
DROP INDEX  IF EXISTS idx_unique_user_group_identifier;
DROP TABLE IF EXISTS "public"."user_group";
DROP SEQUENCE IF EXISTS "public"."id_seq_user_group_mapping";
DROP SEQUENCE IF EXISTS "public"."id_seq_user_group";
ALTER TABLE user_auto_assigned_groups RENAME TO user_groups;