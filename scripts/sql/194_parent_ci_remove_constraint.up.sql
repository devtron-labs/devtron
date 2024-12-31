-- Deleting foreign key relation for parent_ci_pipeline to make it generic, ci_pipeline_type will explain the parent type
-- Handling exists for tables which doesn't have this relation
ALTER TABLE public.ci_pipeline DROP CONSTRAINT IF EXISTS ci_pipeline_parent_ci_pipeline_fkey;