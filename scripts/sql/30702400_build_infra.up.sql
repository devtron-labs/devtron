-- Add the new column `skip_this_value` with default value `false`
ALTER TABLE public.infra_profile_configuration
    ADD COLUMN skip_this_value BOOLEAN NOT NULL DEFAULT FALSE;

-- Update the existing rows where `platform` is "ci-runner" to "default"
UPDATE public.infra_profile_configuration
SET platform = 'default'
WHERE platform = 'ci-runner';


-- Create the sequence for the `id` column if it doesn't already exist
CREATE SEQUENCE IF NOT EXISTS id_seq_profile_platform_mapping;

-- Create the `profile_platform_mapping` table if it doesn't already exist
CREATE TABLE IF NOT EXISTS public.profile_platform_mapping (
    id INTEGER NOT NULL DEFAULT nextval('id_seq_profile_platform_mapping'), -- Primary key with auto-increment
    profile_id INTEGER NOT NULL, -- Foreign key to `infra_profile`
    platform VARCHAR(50) NOT NULL, -- Platform column
    active BOOLEAN NOT NULL, -- Active status column

-- Primary key constraint
    CONSTRAINT profile_platform_mapping_pkey PRIMARY KEY (id),

 -- Foreign key constraint referencing `infra_profile`
    CONSTRAINT fk_profile FOREIGN KEY (profile_id) REFERENCES public.infra_profile (id)
    );


-- Insert data from infra_profile_configuration into profile_platform_mapping
INSERT INTO public.profile_platform_mapping (profile_id, platform, active)
SELECT DISTINCT profile_id, platform, TRUE -- Active is set to TRUE by default
FROM public.infra_profile_configuration
WHERE platform IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM public.profile_platform_mapping ppm
    WHERE ppm.profile_id = infra_profile_configuration.profile_id
      AND ppm.platform = infra_profile_configuration.platform
);