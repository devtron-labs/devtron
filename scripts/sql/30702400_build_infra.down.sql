-- Step 1: Remove the `skip_this_value` column from `infra_profile_configuration`
ALTER TABLE public.infra_profile_configuration
DROP COLUMN IF EXISTS skip_this_value;

-- Step 2: Revert `platform` updates in `infra_profile_configuration` (if needed)
-- Since the original values are not stored, this step can't accurately revert changes.

-- Step 3: Drop the `profile_platform_mapping` table if it exists
DROP TABLE IF EXISTS public.profile_platform_mapping;

-- Step 4: Drop the sequence `id_seq_profile_platform_mapping` if it exists
DROP SEQUENCE IF EXISTS id_seq_profile_platform_mapping;
