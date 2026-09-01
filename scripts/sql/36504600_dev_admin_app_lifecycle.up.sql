-- Dev-Admin (Release 1): decouple application-entity lifecycle from the overloaded create/delete actions.
-- Adds `createApp` / `deleteApp` to the `applications` resource whitelist so custom roles can grant or
-- withhold application create & delete independently of pipeline/workflow/config operations
-- (which remain on `create`/`delete`).
-- Note: jobs are intentionally NOT touched here (job lifecycle is enforced against the jobs resource).
-- Built-in admin/manager/super-admin keep application lifecycle via their wildcard (`*`) action policies
-- and require no change.
UPDATE rbac_policy_resource_detail
SET allowed_actions = array_cat(allowed_actions, ARRAY ['createApp','deleteApp']),
    updated_on      = now()
WHERE resource = 'applications'
  AND allowed_actions IS NOT NULL
  AND NOT ('createApp' = ANY (allowed_actions));