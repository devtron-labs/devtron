-- BEGIN
BEGIN;

ALTER TABLE roles ADD COLUMN IF NOT EXISTS "subaction" VARCHAR(100);

INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('user/entity/accessType', '{
                "value": "user/%/%",
                "indexKeyMap":
                {
                	"5": "Entity",
		            "7": "AccessType",
                }
            }', ARRAY['get','update','create','delete','patch'],'{
                "value": "%/%/%/%/%",
                "indexKeyMap":
                {
                    "0": "TeamObj",
                    "2": "EnvObj",
                    "4": "AppObj",
                    "6": "Action",
                    "8": "SubAction"
                }
            }', ARRAY['apps/devtron-app','apps/helm-app','cluster','jobs','release','chart-group'],'f','now()', 1, 'now()', 1);

ALTER TABLE rbac_role_resource_detail ADD COLUMN IF NOT EXISTS "role_resource_version" VARCHAR(250)[] DEFAULT ARRAY['base']::VARCHAR[];

INSERT INTO rbac_role_resource_detail ("resource", "role_resource_key", "role_resource_update_key",
                                       "eligible_entity_access_types","role_resource_version", "deleted", "created_on", "created_by",
                                       "updated_on", "updated_by")
VALUES ('action', 'Action', 'Action', ARRAY['apps/devtron-app','apps/helm-app','cluster','jobs','release','chart-group'], ARRAY['base','v1'],false, now(), 1, now(), 1);


INSERT INTO rbac_role_resource_detail ("resource", "role_resource_key", "role_resource_update_key",
                                       "eligible_entity_access_types","role_resource_version", "deleted", "created_on", "created_by",
                                       "updated_on", "updated_by")
VALUES ('subAction', 'SubAction', 'SubAction', ARRAY['apps/devtron-app','apps/helm-app','cluster','jobs','release','chart-group'], ARRAY['base','v1'],false, now(), 1, now(), 1);

COMMIT ;