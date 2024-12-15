-- Add the new column `skip_this_value` with default value `false` currently not needed
-- ALTER TABLE public.infra_profile_configuration
--     ADD COLUMN IF NOT EXISTS skip_this_value BOOLEAN NOT NULL DEFAULT FALSE;


-- Begin Transaction to ensure atomicity of all operations
BEGIN;

-- ---------------------------------------------------
-- Step 1: Modify infra_profile_configuration Table
-- ---------------------------------------------------

-- Optionally add the `skip_this_value` column if needed in the future
-- ALTER TABLE public.infra_profile_configuration
--     ADD COLUMN IF NOT EXISTS skip_this_value BOOLEAN NOT NULL DEFAULT FALSE;


-- Ensure no NULL values exist before setting NOT NULL
UPDATE public.infra_profile_configuration
SET platform = 'default';

-- Alter the `platform` column to set NOT NULL and default to 'default'
ALTER TABLE public.infra_profile_configuration
    ALTER COLUMN platform SET NOT NULL,
--     ALTER COLUMN platform SET DEFAULT 'default';

-- Remove the default value from the `platform` column to enforce explicit assignment in the future
-- ALTER TABLE public.infra_profile_configuration
--     ALTER COLUMN platform DROP DEFAULT;


-- ---------------------------------------------------
-- Step 2: Update Existing Data in infra_profile_configuration
-- ---------------------------------------------------

-- Update rows where `platform` is 'ci-runner' to 'default'
UPDATE public.infra_profile_configuration
SET platform = 'default'
WHERE platform = 'ci-runner';


-- ---------------------------------------------------
-- Step 3: Update infra_profile Table
-- ---------------------------------------------------

-- Update the `name` from 'default' to 'global' in infra_profile
UPDATE public.infra_profile
SET name = 'global'
WHERE name = 'default';

-- ---------------------------------------------------
-- Step 4: Create Sequence for profile_platform_mapping
-- ---------------------------------------------------

CREATE SEQUENCE IF NOT EXISTS public.id_seq_profile_platform_mapping;

-- ---------------------------------------------------
-- Step 5: Create profile_platform_mapping Table
-- ---------------------------------------------------
CREATE TABLE IF NOT EXISTS public.profile_platform_mapping (
    id INT4 NOT NULL DEFAULT nextval('public.id_seq_profile_platform_mapping'), -- Primary key with auto-increment
    profile_id INTEGER NOT NULL, -- Foreign key to `infra_profile`
    platform VARCHAR(50) NOT NULL, -- Platform column
    active BOOLEAN NOT NULL, -- Active status column
    created_by INTEGER NOT NULL, -- Who created the record
    updated_on TIMESTAMPTZ NOT NULL, -- Last updated timestamp
    updated_by INTEGER NOT NULL, -- Who last updated the record
    created_on TIMESTAMPTZ NOT NULL, -- Timestamp of record creation

-- Primary key constraint
    CONSTRAINT profile_platform_mapping_pkey PRIMARY KEY (id),

    -- Foreign key constraint referencing `infra_profile`
    CONSTRAINT fk_profile FOREIGN KEY (profile_id) REFERENCES public.infra_profile (id) ON DELETE CASCADE
    );



-- ---------------------------------------------------
-- Step 6: Insert Default Platform Mappings from infra_profile
-- ---------------------------------------------------
-- Insert Default Platform Mappings from infra_profile
INSERT INTO public.profile_platform_mapping (
    profile_id, platform, active, created_by, updated_on, updated_by, created_on
)
SELECT DISTINCT
    ip.id AS profile_id,
    'default' AS platform,
    TRUE AS active,
    1 AS created_by, -- Replace with appropriate user ID or parameter
    CURRENT_TIMESTAMP AS updated_on,
    1 AS updated_by, -- Replace with appropriate user ID or parameter
    CURRENT_TIMESTAMP AS created_on
FROM public.infra_profile ip
WHERE ip.active IS TRUE
  AND NOT EXISTS (
    SELECT 1
    FROM public.profile_platform_mapping ppm
    WHERE ppm.profile_id = ip.id
      AND ppm.platform = 'default'
);

-- ---------------------------------------------------
-- Step 7: Insert Platform Mappings from infra_profile_configuration
-- ---------------------------------------------------
-- Insert Platform Mappings from infra_profile_configuration
INSERT INTO public.profile_platform_mapping (
    profile_id, platform, active, created_by, updated_on, updated_by, created_on
)
SELECT DISTINCT
    ipc.profile_id,
    ipc.platform,
    TRUE AS active,
    1 AS created_by, -- Replace with appropriate user ID or parameter
    CURRENT_TIMESTAMP AS updated_on,
    1 AS updated_by, -- Replace with appropriate user ID or parameter
    CURRENT_TIMESTAMP AS created_on
FROM public.infra_profile_configuration ipc
WHERE  ipc.active IS TRUE AND ipc.platform IS NOT NULL
  AND NOT EXISTS (
    SELECT 1
    FROM public.profile_platform_mapping ppm
    WHERE ppm.profile_id = ipc.profile_id
      AND ppm.platform = ipc.platform
);
-- ---------------------------------------------------
-- Commit Transaction
-- ---------------------------------------------------
COMMIT;
