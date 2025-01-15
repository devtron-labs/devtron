BEGIN;




ALTER TABLE public.infra_profile
    DROP COLUMN IF EXISTS buildx_driver_type;

ALTER TABLE public.infra_profile_configuration
    DROP CONSTRAINT IF EXISTS fk_infra_profile_configuration_profile_platform_mapping_id;

ALTER TABLE public.infra_profile_configuration
    DROP COLUMN IF EXISTS profile_platform_mapping_id;

DROP  INDEX IF EXISTS unique_profile_platform_mapping;

END;