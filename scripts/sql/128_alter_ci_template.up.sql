ALTER ci_template ADD COLUMN build_context VARCHAR(200) DEFAULT '.'
ALTER ci_template ADD COLUMN build_context_git_material_id INT;
UPDATE ci_template SET build_context_git_material_id = git_material_id;