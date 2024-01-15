DELETE FROM plugin_stage_mapping where plugin_id=(SELECT id from plugin_metadata where name='DockerSlim v1.1.0');
DELETE FROM plugin_step where plugin_id=(SELECT id FROM plugin_metadata WHERE name='DockerSlim v1.1.0');
DELETE FROM plugin_step_variable where plugin_step_id=(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='DockerSlim v1.1.0' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_metadata where name='DockerSlim v1.1.0';