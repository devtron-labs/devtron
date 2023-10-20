UPDATE "public"."rbac_policy_resource_detail" set
eligible_entity_access_types = ARRAY['apps/devtron-app'] where resource='applications' OR resource ='user';

UPDATE "public"."rbac_policy_resource_detail" set
eligible_entity_access_types = ARRAY['apps/devtron-app','apps/helm-app'] where resource='project' OR resource ='global-environment' OR resource='terminal';

UPDATE "public"."rbac_role_resource_detail" set
eligible_entity_access_types = ARRAY['apps/devtron-app','apps/helm-app'] where resource='applications' OR resource ='project' OR resource ='environment';

DELETE FROM "public"."rbac_policy_resource_detail" WHERE resource ='appEnv';

DELETE FROM "public"."rbac_role_data" WHERE entity='apps' AND access_type ='jobs';

DELETE FROM "public"."rbac_policy_data" WHERE entity='apps' AND access_type ='jobs';
