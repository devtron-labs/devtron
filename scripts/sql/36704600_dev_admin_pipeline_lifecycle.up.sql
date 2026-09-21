-- Increment 2: add createPipeline / deletePipeline to the applications resource whitelist so custom
-- roles can grant/withhold pipeline lifecycle independently of config actions. dev-admin omits them.
UPDATE rbac_policy_resource_detail
SET allowed_actions = array_cat(allowed_actions, ARRAY ['createPipeline','deletePipeline']),
    updated_on      = now()
WHERE resource = 'applications'
  AND allowed_actions IS NOT NULL
  AND NOT ('createPipeline' = ANY (allowed_actions));
