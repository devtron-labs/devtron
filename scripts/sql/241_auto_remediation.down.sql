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
DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Custom Notifier v1.0.0' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Custom Notifier v1.0.0');
DELETE FROM plugin_stage_mapping WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE name='Custom Notifier v1.0.0');
DELETE FROM pipeline_stage_step_variable WHERE pipeline_stage_step_id in (SELECT id FROM pipeline_stage_step where ref_plugin_id =(SELECT id from plugin_metadata WHERE name ='Custom Notifier v1.0.0'));
DELETE FROM pipeline_stage_step where ref_plugin_id in (SELECT id from plugin_metadata WHERE name ='Custom Notifier v1.0.0');
DELETE FROM plugin_metadata WHERE name ='Custom Notifier v1.0.0';