DELETE FROM plugin_step_variable WHERE plugin_step_id in (SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.plugin_version='1.0.0' and p.name='Docker Lint' and p.deleted=false and ps."index"=1 and ps.deleted=false);


DELETE FROM plugin_step WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='Docker Lint' and deleted=false);


DELETE FROM plugin_stage_mapping WHERE plugin_id =(SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='Docker Lint' and deleted=false);


DELETE FROM plugin_tag_relation WHERE plugin_id in (SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='Docker Lint' and deleted=false);

DELETE FROM pipeline_stage_step_variable where pipeline_stage_step_id in (select id from pipeline_stage_step where name = 'Docker Lint') ;

DELETE FROM pipeline_stage_step where ref_plugin_id in (SELECT id from plugin_metadata WHERE plugin_version='1.0.0' and name ='Docker Lint' and deleted=false);


DELETE from plugin_pipeline_script where id = (SELECT script_id from plugin_step WHERE plugin_id=(SELECT id FROM plugin_metadata WHERE plugin_version='1.0.0' and name='Docker Lint' and deleted=false));


DELETE FROM plugin_metadata WHERE plugin_version='1.0.0' and name ='Docker Lint' and deleted=false;


DELETE FROM plugin_parent_metadata WHERE identifier ='docker-lint';