BEGIN;

UPDATE "public"."infra_profile_configuration"
SET active='f'
WHERE platform = 'runner'
  AND value_string= '[]'
  AND active='t'
  AND profile_id=(
    SELECT infra_Profile.id
    FROM infra_profile
    WHERE name = 'global'
      AND active='t'
)
;

ALTER TABLE public.infra_profile_configuration
    ALTER COLUMN platform DROP NOT NULL;

ALTER TABLE public.infra_profile
    ADD COLUMN IF NOT EXISTS buildx_driver_type VARCHAR(50) NOT NULL DEFAULT 'kubernetes';

ALTER TABLE public.infra_profile_configuration
    ADD COLUMN IF NOT EXISTS profile_platform_mapping_id INTEGER;

ALTER TABLE public.infra_profile_configuration
    DROP CONSTRAINT IF EXISTS fk_infra_profile_configuration_profile_platform_mapping_id;

ALTER TABLE public.infra_profile_configuration
    ADD CONSTRAINT fk_infra_profile_configuration_profile_platform_mapping_id
        FOREIGN KEY (profile_platform_mapping_id) REFERENCES profile_platform_mapping(id);

CREATE UNIQUE INDEX IF NOT EXISTS unique_profile_platform_mapping ON public.profile_platform_mapping
    (profile_id, platform) where active=true;

UPDATE infra_profile_configuration
    SET profile_platform_mapping_id = profile_platform_mapping.id
    FROM profile_platform_mapping
    WHERE infra_profile_configuration.profile_id = profile_platform_mapping.profile_id
        AND infra_profile_configuration.platform = profile_platform_mapping.platform
        AND profile_platform_mapping.active = 't'
        AND infra_profile_configuration.active = 't';

END;