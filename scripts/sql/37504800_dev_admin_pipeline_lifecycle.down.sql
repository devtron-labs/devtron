-- Reverse Increment 2: remove pipeline lifecycle actions and restore Release-1 lifecycle actions to
-- their original mixed-case form (as 36504600 left them), so down returns to the pre-migration state.

-- 1) remove pipeline lifecycle actions
UPDATE rbac_policy_resource_detail
SET allowed_actions = array_remove(array_remove(allowed_actions, 'createpipeline'), 'deletepipeline'),
    updated_on      = now()
WHERE resource = 'applications';

-- 2) restore Release-1 application lifecycle actions to mixed case
UPDATE rbac_policy_resource_detail
SET allowed_actions = array_replace(array_replace(allowed_actions, 'createapp', 'createApp'), 'deleteapp', 'deleteApp'),
    updated_on      = now()
WHERE resource = 'applications'
  AND allowed_actions IS NOT NULL;
