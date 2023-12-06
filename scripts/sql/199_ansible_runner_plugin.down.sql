DELETE FROM plugin_step_variable WHERE plugin_step_id =(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Ansible Runner' and ps."index"=1 and ps.deleted=false);
DELETE FROM plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE name='Ansible Runner');
DELETE FROM plugin_stage_mapping WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE name='Ansible Runner');
DELETE from pipeline_stage_step_variable where pipeline_stage_step_id =(select pipeline_stage_id from pipeline_stage_step where name='Ansible Runner');
DELETE from pipeline_stage_step where name ='Ansible Runner';
DELETE FROM plugin_metadata WHERE name ='Ansible Runner';
