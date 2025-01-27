
BEGIN;

ALTER TABLE public.infra_profile
  ALTER COLUMN description TYPE VARCHAR(300);

UPDATE public.infra_profile
SET name = 'default'
WHERE name = 'global';

-- Step 2: Modify infra_profile_configuration Table

-- 2.1: Remove NOT NULL Constraint from 'platform' Column
ALTER TABLE public.infra_profile_configuration
    ALTER COLUMN platform DROP NOT NULL;

-- ---------------------------------------------------
-- Step 3: Revert Data Changes in infra_profile_configuration
-- ---------------------------------------------------
UPDATE public.infra_profile_configuration
SET platform = 'ci-runner'
WHERE platform = 'runner';

-- ---------------------------------------------------
-- Step 4: Drop profile_platform_mapping Table
-- ---------------------------------------------------

-- This will automatically drop any indexes and constraints associated with the table.
DROP TABLE IF EXISTS public.profile_platform_mapping;

-- ---------------------------------------------------
-- Step 5: Drop Sequence for profile_platform_mapping
-- ---------------------------------------------------

DROP SEQUENCE IF EXISTS public.id_seq_profile_platform_mapping;
-- ---------------------------------------------------
-- Step 6: Commit Transaction
-- ---------------------------------------------------
COMMIT;
