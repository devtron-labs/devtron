-- Drop indexes
DROP INDEX IF EXISTS idx_unique_k8s_event_watcher_name;

-- Drop foreign key constraints
ALTER TABLE "public"."intercepted_event_execution" DROP CONSTRAINT IF EXISTS intercepted_events_auto_remediation_trigger_id_fkey;
ALTER TABLE "public"."intercepted_event_execution" DROP CONSTRAINT IF EXISTS intercepted_events_cluster_id_fkey;
ALTER TABLE "public"."auto_remediation_trigger" DROP CONSTRAINT IF EXISTS auto_remediation_trigger_k8s_event_watcher_id_fkey;

-- Drop tables
DROP TABLE IF EXISTS "public"."intercepted_event_execution";
DROP TABLE IF EXISTS "public"."auto_remediation_trigger";
DROP TABLE IF EXISTS "public"."k8s_event_watcher";

-- Drop sequences
DROP SEQUENCE IF EXISTS id_seq_intercepted_events;
DROP SEQUENCE IF EXISTS id_seq_auto_remediation_trigger;
DROP SEQUENCE IF EXISTS id_seq_k8s_event_watcher;


-- PLUGIN DOWN SCRIPT
DELETE FROM plugin_step_variable WHERE plugin_step_id=(SELECT id FROM plugin_metadata WHERE name='Custom Notification v1.0.0');
DELETE FROM plugin_step where plugin_id=(SELECT id FROM plugin_metadata WHERE name='Custom Notification v1.0.0' );
DELETE FROM plugin_pipeline_script where id=(SELECT id FROM plugin_metadata WHERE name='Custom Notification v1.0.0');
DELETE FROM plugin_stage_mapping where plugin_id=(SELECT id from plugin_metadata where name='Custom Notification v1.0.0');
DELETE FROM plugin_metadata where name='Custom Notification v1.0.0';