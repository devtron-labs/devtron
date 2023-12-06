DELETE from plugin_step_variable where plugin_step_id=(SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Terraform CLI' and ps."index"=1 and ps.deleted=false);
DELETE from plugin_step where plugin_id=(select id from plugin_metadata where name='Terraform CLI');
DELETE from plugin_stage_mapping where plugin_id=(SELECT id from plugin_metadata where name='Terraform CLI');
DELETE from pipeline_stage_step_variable where pipeline_stage_step_id =(select pipeline_stage_id from pipeline_stage_step where name='Terraform CLI');
DELETE from pipeline_stage_step where name ='Terraform CLI';
DELETE from plugin_metadata where name='Terraform CLI';
