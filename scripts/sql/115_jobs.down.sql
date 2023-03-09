DELETE FROM ci_pipeline_material WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2));

DELETE FROM ci_template_history WHERE git_material_id IN (SELECT id FROM git_material_history WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2)));

DELETE FROM ci_pipeline_history WHERE ci_pipeline_id IN (SELECT id FROM ci_pipeline WHERE ci_template_id IN (SELECT id FROM ci_template WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2))));

DELETE FROM ci_artifact WHERE ci_workflow_id IN (SELECT id FROM ci_workflow WHERE ci_pipeline_id IN (SELECT id FROM ci_pipeline WHERE ci_template_id IN (SELECT id FROM ci_template WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2)))));

DELETE FROM ci_workflow WHERE ci_pipeline_id IN (SELECT id FROM ci_pipeline WHERE ci_template_id IN (SELECT id FROM ci_template WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2))));

DELETE FROM pipeline_stage WHERE ci_pipeline_id IN (SELECT id FROM ci_pipeline WHERE ci_template_id IN (SELECT id FROM ci_template WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2))));

DELETE FROM ci_pipeline WHERE ci_template_id IN (SELECT id FROM ci_template WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2)));

DELETE FROM ci_workflow WHERE ci_pipeline_id IN (SELECT id FROM ci_pipeline WHERE ci_template_id IN (SELECT id FROM ci_template WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2))));

DELETE FROM ci_template_history WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2));

DELETE FROM ci_template WHERE git_material_id IN (SELECT id FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2));

DELETE FROM git_material_history WHERE git_material_id IN(SELECT id FROM git_material WHERE app_id IN(SELECT id FROM app WHERE app_store = 2));

DELETE FROM git_material WHERE app_id IN (SELECT id FROM app WHERE app_store = 2);

DELETE FROM app_label WHERE app_id IN (SELECT id FROM app WHERE app_store = 2);

DELETE FROM app_workflow_mapping WHERE app_workflow_id IN(SELECT id FROM app_workflow WHERE app_id IN (SELECT id FROM app WHERE app_store = 2));

DELETE FROM app_workflow WHERE app_id IN (SELECT id FROM app WHERE app_store = 2);

DELETE FROM app WHERE app_store = 2;

ALTER TABLE app ALTER COLUMN app_store DROP DEFAULT;

ALTER TABLE app ALTER app_store TYPE boolean USING CASE WHEN app_store=1 THEN true ELSE false end;

ALTER TABLE app ALTER COLUMN app_store SET DEFAULT FALSE;

ALTER TABLE app DROP COLUMN display_name;

ALTER TABLE app DROP COLUMN description;

ALTER TABLE ci_artifact DROP COLUMN is_artifact_uploaded;