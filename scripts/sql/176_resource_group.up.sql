
ALTER TABLE public.app_group_mapping DROP CONSTRAINT app_group_mapping_app_id_fkey;
ALTER TABLE public.app_group_mapping DROP CONSTRAINT app_group_mapping_app_group_id_fkey;
ALTER TABLE public.app_group_mapping RENAME COLUMN app_group_id TO resource_group_id;
ALTER TABLE public.app_group_mapping RENAME COLUMN app_id TO resource_id;
ALTER TABLE public.app_group_mapping ADD resource_key int4 NOT NULL DEFAULT 6;
ALTER TABLE public.app_group_mapping RENAME TO resource_group_mapping;


ALTER TABLE public.app_group RENAME TO resource_group;
ALTER TABLE public.resource_group RENAME COLUMN environment_id TO resource_id;
ALTER TABLE public.resource_group ADD resource_key int4 NOT NULL DEFAULT 7;


ALTER TABLE public.resource_group_mapping ADD CONSTRAINT resource_group_mapping_fk FOREIGN KEY (resource_group_id) REFERENCES public.resource_group(id);