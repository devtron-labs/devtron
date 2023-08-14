DELETE FROM "public"."plugin_tag_relation" WHERE plugin_id in ((SELECT id FROM plugin_metadata WHERE name='Sonarqube PR Based'));

DELETE FROM "public"."plugin_pipeline_script" WHERE (SELECT ps.script_id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based');

DELETE FROM "public"."plugin_step_variable" WHERE plugin_step_id in ((SELECT ps.id FROM plugin_metadata p inner JOIN plugin_step ps on ps.plugin_id=p.id WHERE p.name='Sonarqube PR Based' and ps."index"=1 and ps.deleted=false));

DELETE FROM "public"."plugin_step" WHERE plugin_id in ((SELECT id FROM plugin_metadata WHERE name='Sonarqube PR Based'));

DELETE FROM "public"."plugin_stage_mapping" ("plugin_id" WHERE plugin_id = (SELECT id FROM plugin_metadata WHERE name='Sonarqube PR Based'));

DELETE FROM "public"."plugin_metadata" WHERE name = 'Sonarqube PR Based';