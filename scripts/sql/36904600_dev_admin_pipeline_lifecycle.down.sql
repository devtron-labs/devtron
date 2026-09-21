UPDATE rbac_policy_resource_detail
SET allowed_actions = array_remove(array_remove(allowed_actions, 'createPipeline'), 'deletePipeline'),
    updated_on      = now()
WHERE resource = 'applications';
