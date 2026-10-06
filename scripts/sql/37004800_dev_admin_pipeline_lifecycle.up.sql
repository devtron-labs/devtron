-- Increment 2: add createpipeline / deletepipeline to the applications resource whitelist so custom
-- roles can grant/withhold pipeline lifecycle independently of config actions. dev-admin omits them.
--
-- Action values are lowercase on purpose: the casbin store normalizes policy actions to lowercase on
-- AddPolicy (strings.ToLower) while enforcement passes the action unchanged, so a mixed-case whitelist
-- label would never match the stored grant. This migration also normalizes the Release-1 lifecycle
-- actions (createApp/deleteApp, added by 36504600) to lowercase here rather than editing that already
-- shipped migration, so the whitelist is internally consistent on both fresh and upgraded installs.

-- 1) normalize Release-1 application lifecycle actions to lowercase (no-op if already lowercase)
UPDATE rbac_policy_resource_detail
SET allowed_actions = array_replace(array_replace(allowed_actions, 'createApp', 'createapp'), 'deleteApp', 'deleteapp'),
    updated_on      = now()
WHERE resource = 'applications'
  AND allowed_actions IS NOT NULL;

-- 2) add pipeline lifecycle actions (lowercase)
UPDATE rbac_policy_resource_detail
SET allowed_actions = array_cat(allowed_actions, ARRAY ['createpipeline','deletepipeline']),
    updated_on      = now()
WHERE resource = 'applications'
  AND allowed_actions IS NOT NULL
  AND NOT ('createpipeline' = ANY (allowed_actions));
