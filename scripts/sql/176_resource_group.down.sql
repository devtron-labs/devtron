
-- Reverting changes made to app_group_mapping
ALTER TABLE public.resource_group_mapping DROP CONSTRAINT resource_group_mapping_fk;
ALTER TABLE public.resource_group_mapping RENAME TO app_group_mapping;
ALTER TABLE public.app_group_mapping DROP COLUMN resource_key;
ALTER TABLE public.app_group_mapping RENAME COLUMN resource_id TO app_id;
ALTER TABLE public.app_group_mapping RENAME COLUMN resource_group_id TO app_group_id;


-- Reverting changes made to resource_group
ALTER TABLE public.resource_group DROP COLUMN resource_key;
ALTER TABLE public.resource_group RENAME COLUMN resource_id TO environment_id;
ALTER TABLE public.resource_group RENAME TO app_group;


ALTER TABLE public.app_group_mapping ADD CONSTRAINT app_group_mapping_app_group_id_fkey FOREIGN KEY (app_group_id) REFERENCES public.app_group(id);
ALTER TABLE public.app_group_mapping ADD CONSTRAINT app_group_mapping_app_id_fkey FOREIGN KEY (app_id) REFERENCES public.app(id);
