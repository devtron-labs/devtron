ALTER TABLE public.git_material_history
    DROP CONSTRAINT IF EXISTS git_material_history_cloning_mode_check,
    DROP COLUMN IF EXISTS cloning_mode;

ALTER TABLE public.git_material
    DROP CONSTRAINT IF EXISTS git_material_cloning_mode_check,
    DROP COLUMN IF EXISTS cloning_mode;
