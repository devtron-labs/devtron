ALTER TABLE app_workflow  DROP CONSTRAINT app_workflow_app_id_fkey,  ADD CONSTRAINT app_workflow_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE;

ALTER TABLE ci_pipeline DROP CONSTRAINT ci_pipeline_app_id_fkey, ADD CONSTRAINT ci_pipeline_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE;

ALTER TABLE ci_template DROP CONSTRAINT ci_template_app_id_fkey, ADD CONSTRAINT ci_template_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE;

ALTER TABLE git_material DROP CONSTRAINT git_material_app_id_fkey, ADD CONSTRAINT git_material_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE;

ALTER TABLE app_label DROP CONSTRAINT app_label_app_id_fkey, ADD CONSTRAINT app_label_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE;

ALTER TABLE ci_template_history DROP CONSTRAINT ci_template_history_app_id_fkey, ADD CONSTRAINT ci_template_history_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id) ON DELETE CASCADE;

ALTER TABLE app_workflow_mapping DROP CONSTRAINT app_workflow_mapping_app_workflow_id_fkey, ADD CONSTRAINT app_workflow_mapping_app_workflow_id_fkey FOREIGN KEY (app_workflow_id) REFERENCES app_workflow(id) ON DELETE CASCADE;

ALTER TABLE ci_pipeline_material DROP CONSTRAINT ci_pipeline_material_ci_pipeline_id_fkey, ADD CONSTRAINT ci_pipeline_material_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES  ci_pipeline(id) ON DELETE CASCADE;

ALTER TABLE ci_pipeline_history DROP CONSTRAINT ci_pipeline_history_ci_pipeline_id_fk, ADD CONSTRAINT ci_pipeline_history_ci_pipeline_id_fk FOREIGN KEY (ci_pipeline_id) REFERENCES  ci_pipeline(id) ON DELETE CASCADE;

ALTER TABLE git_material_history DROP CONSTRAINT git_material_history_git_material_id_fkey, ADD CONSTRAINT git_material_history_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES   git_material(id) ON DELETE CASCADE;

ALTER TABLE ci_workflow DROP CONSTRAINT ci_workflow_ci_pipeline_id_fkey, ADD CONSTRAINT ci_workflow_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES   ci_pipeline(id) ON DELETE CASCADE;

ALTER TABLE ci_pipeline_material DROP CONSTRAINT ci_pipeline_material_git_material_id_fkey, ADD CONSTRAINT ci_pipeline_material_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES  git_material(id) ON DELETE CASCADE;

ALTER TABLE ci_template DROP CONSTRAINT ci_template_git_material_id_fkey, ADD CONSTRAINT ci_template_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES  git_material(id) ON DELETE CASCADE;

ALTER TABLE ci_template_history DROP CONSTRAINT ci_template_git_material_history_id_fkey, ADD CONSTRAINT ci_template_git_material_history_id_fkey FOREIGN KEY (git_material_id) REFERENCES  git_material(id) ON DELETE CASCADE;

ALTER TABLE ci_pipeline DROP CONSTRAINT ci_pipeline_ci_template_id_fkey, ADD CONSTRAINT ci_pipeline_ci_template_id_fkey FOREIGN KEY (ci_template_id) REFERENCES  ci_template(id) ON DELETE CASCADE;

DELETE FROM app WHERE app_store = 2;

ALTER TABLE app ALTER COLUMN app_store DROP DEFAULT;

ALTER TABLE app ALTER app_store TYPE boolean USING CASE WHEN app_store=1 THEN true ELSE false end;

ALTER TABLE app ALTER COLUMN app_store SET DEFAULT FALSE;

ALTER TABLE app DROP COLUMN display_name;

ALTER TABLE app DROP COLUMN description;

ALTER TABLE ci_pipeline DROP CONSTRAINT ci_pipeline_ci_template_id_fkey;

ALTER TABLE ci_pipeline ADD CONSTRAINT ci_pipeline_ci_template_id_fkey FOREIGN KEY (ci_template_id) REFERENCES ci_template(id);

ALTER TABLE ci_template_history DROP CONSTRAINT ci_template_git_material_history_id_fkey;

ALTER TABLE ci_template_history ADD CONSTRAINT ci_template_git_material_history_id_fkey FOREIGN KEY (git_material_id) REFERENCES git_material(id);

ALTER TABLE ci_template DROP CONSTRAINT ci_template_git_material_id_fkey;

ALTER TABLE ci_template ADD CONSTRAINT ci_template_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES git_material(id);

ALTER TABLE ci_pipeline_material DROP CONSTRAINT ci_pipeline_material_git_material_id_fkey;

ALTER TABLE ci_pipeline_material ADD CONSTRAINT ci_pipeline_material_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES git_material(id);

ALTER TABLE ci_workflow DROP CONSTRAINT ci_workflow_ci_pipeline_id_fkey;

ALTER TABLE ci_workflow ADD CONSTRAINT ci_workflow_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES ci_pipeline(id);

ALTER TABLE git_material_history DROP CONSTRAINT git_material_history_git_material_id_fkey;

ALTER TABLE git_material_history ADD CONSTRAINT git_material_history_git_material_id_fkey FOREIGN KEY (git_material_id) REFERENCES git_material(id);

ALTER TABLE ci_pipeline_history DROP CONSTRAINT ci_pipeline_history_ci_pipeline_id_fk;

ALTER TABLE ci_pipeline_history ADD CONSTRAINT ci_pipeline_history_ci_pipeline_id_fk FOREIGN KEY (ci_pipeline_id) REFERENCES ci_pipeline(id);

ALTER TABLE ci_pipeline_material DROP CONSTRAINT ci_pipeline_material_ci_pipeline_id_fkey;

ALTER TABLE ci_pipeline_material ADD CONSTRAINT ci_pipeline_material_ci_pipeline_id_fkey FOREIGN KEY (ci_pipeline_id) REFERENCES ci_pipeline(id);

ALTER TABLE app_workflow_mapping DROP CONSTRAINT app_workflow_mapping_app_workflow_id_fkey;

ALTER TABLE app_workflow_mapping ADD CONSTRAINT app_workflow_mapping_app_workflow_id_fkey FOREIGN KEY (app_workflow_id) REFERENCES app_workflow(id);

ALTER TABLE ci_template_history DROP CONSTRAINT ci_template_history_app_id_fkey;

ALTER TABLE ci_template_history ADD CONSTRAINT ci_template_history_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id);

ALTER TABLE app_label DROP CONSTRAINT app_label_app_id_fkey;

ALTER TABLE app_label ADD CONSTRAINT app_label_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id);

ALTER TABLE git_material DROP CONSTRAINT git_material_app_id_fkey;

ALTER TABLE git_material ADD CONSTRAINT git_material_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id);

ALTER TABLE ci_template DROP CONSTRAINT ci_template_app_id_fkey;

ALTER TABLE git_material ADD CONSTRAINT ci_template_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id);

ALTER TABLE ci_pipeline DROP CONSTRAINT ci_pipeline_app_id_fkey;

ALTER TABLE ci_pipeline ADD CONSTRAINT ci_pipeline_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id);

ALTER TABLE app_workflow DROP CONSTRAINT app_workflow_app_id_fkey;

ALTER TABLE app_workflow ADD CONSTRAINT app_workflow_app_id_fkey FOREIGN KEY (app_id) REFERENCES app(id);