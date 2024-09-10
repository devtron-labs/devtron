UPDATE ci_pipeline_material SET  active = false FROM git_material JOIN app ON app.id = git_material.app_id WHERE app.app_type = 2 AND ci_pipeline_material.git_material_id = git_material.id;

UPDATE ci_pipeline SET active = false WHERE app_id IN (SELECT id FROM app WHERE app_type = 2);

UPDATE ci_template SET active = false WHERE app_id IN (SELECT id FROM app WHERE app_type = 2);

UPDATE git_material SET active = false WHERE app_id IN (SELECT id FROM app WHERE app_type = 2);

UPDATE app_workflow_mapping SET active = false FROM app_workflow JOIN app ON app_workflow.app_id = app.id WHERE app.app_type = 2 AND app_workflow_mapping.app_workflow_id = app_workflow.id;

UPDATE app_workflow SET active = false WHERE app_id IN (SELECT id FROM app WHERE app_type = 2);

UPDATE app SET active = false WHERE app_type = 2;



