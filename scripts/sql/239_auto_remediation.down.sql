-- Drop indexes
DROP INDEX IF EXISTS idx_unique_watcher_name;

-- Drop foreign key constraints
ALTER TABLE "public"."intercepted_event_execution" DROP CONSTRAINT IF EXISTS intercepted_events_trigger_id_fkey;
ALTER TABLE "public"."intercepted_event_execution" DROP CONSTRAINT IF EXISTS intercepted_events_cluster_id_fkey;
ALTER TABLE "public"."trigger" DROP CONSTRAINT IF EXISTS trigger_watcher_id_fkey;

-- Drop tables
DROP TABLE IF EXISTS "public"."intercepted_event_execution";
DROP TABLE IF EXISTS "public"."trigger";
DROP TABLE IF EXISTS "public"."watcher";

-- Drop sequences
DROP SEQUENCE IF EXISTS id_seq_intercepted_events;
DROP SEQUENCE IF EXISTS id_seq_trigger;
DROP SEQUENCE IF EXISTS id_seq_watcher;
