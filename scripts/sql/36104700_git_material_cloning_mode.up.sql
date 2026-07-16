ALTER TABLE public.git_material
    ADD COLUMN cloning_mode VARCHAR(16) NOT NULL DEFAULT 'FULL',
    ADD CONSTRAINT git_material_cloning_mode_check CHECK (cloning_mode IN ('FULL', 'SHALLOW'));

ALTER TABLE public.git_material_history
    ADD COLUMN cloning_mode VARCHAR(16) NOT NULL DEFAULT 'FULL',
    ADD CONSTRAINT git_material_history_cloning_mode_check CHECK (cloning_mode IN ('FULL', 'SHALLOW'));
