ALTER TABLE ci_template DROP COLUMN IF EXISTS build_context;
ALTER TABLE ci_template DROP COLUMN IF EXISTS build_context_git_material_id;
ALTER TABLE ci_template DROP CONSTRAINT IF EXISTS ci_template_build_context_git_material_id_fkey;
ALTER TABLE ci_template_override DROP COLUMN IF EXISTS build_context_git_material_id;
ALTER TABLE ci_template_override DROP CONSTRAINT IF EXISTS ci_template_override_build_context_git_material_id_fkey;
