ALTER TABLE ci_template ADD COLUMN build_context_git_material_id INT;
UPDATE ci_template SET build_context_git_material_id = git_material_id;
ALTER TABLE ci_template_override ADD COLUMN build_context_git_material_id INT;
UPDATE ci_template_override SET build_context_git_material_id = git_material_id;

ALTER TABLE  ci_template
    ADD CONSTRAINT "ci_template_build_context_git_material_id_fkey" FOREIGN KEY ("build_context_git_material_id") REFERENCES "public"."git_material" ("id");

ALTER TABLE  ci_template_override
    ADD CONSTRAINT "ci_template_override_build_context_git_material_id_fkey" FOREIGN KEY ("build_context_git_material_id") REFERENCES "public"."git_material" ("id");