ALTER TABLE ci_template ADD COLUMN IF NOT EXISTS use_root_build_context bool DEFAULT true;
ALTER TABLE ci_template_override ADD COLUMN IF NOT EXISTS use_root_build_context bool DEFAULT true;