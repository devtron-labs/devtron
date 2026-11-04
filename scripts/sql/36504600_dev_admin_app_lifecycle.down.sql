-- Revert: remove `createApp` / `deleteApp` from the `applications` resource whitelist.
UPDATE rbac_policy_resource_detail
SET allowed_actions = array_remove(array_remove(allowed_actions, 'createApp'), 'deleteApp'),
    updated_on      = now()
WHERE resource = 'applications';