
INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('release', '{"value": "release", "indexKeyMap": {}}', ARRAY['get','update','create','delete','patch'],'{"value": "%/%", "indexKeyMap": {"0": "ReleaseTrackObj", "2": "ReleaseObj"}}', ARRAY['release'],'f','now()', 1, 'now()', 1);

INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('release-requirement', '{"value": "release-requirement", "indexKeyMap": {}}', ARRAY['get','update','create','delete','patch'],'{"value": "%/%", "indexKeyMap": {"0": "ReleaseTrackObj", "2": "ReleaseObj"}}', ARRAY['release'],'f','now()', 1, 'now()', 1);

INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('release-track', '{"value": "release-track", "indexKeyMap": {}}', ARRAY['get','update','create','delete','patch'],'{"value": "%", "indexKeyMap": {"0": "ReleaseTrackObj"}}', ARRAY['release'],'f','now()', 1, 'now()', 1);

INSERT INTO rbac_policy_resource_detail ("resource", "policy_resource_value", "allowed_actions",
                                         "resource_object", "eligible_entity_access_types", "deleted", "created_on",
                                         "created_by", "updated_on", "updated_by")
VALUES ('release-track-requirement', '{"value": "release-track-requirement", "indexKeyMap": {}}', ARRAY['get','update','create','delete','patch'],'{"value": "%", "indexKeyMap": {"0": "ReleaseTrackObj"}}', ARRAY['release'],'f','now()', 1, 'now()', 1);





INSERT INTO rbac_role_resource_detail ("resource", "role_resource_key", "role_resource_update_key",
                                       "eligible_entity_access_types", "deleted", "created_on", "created_by",
                                       "updated_on", "updated_by")
VALUES ('release', 'Release', 'Release', ARRAY ['release'], false, now(), 1, now(), 1);


INSERT INTO rbac_role_resource_detail ("resource", "role_resource_key", "role_resource_update_key",
                                       "eligible_entity_access_types", "deleted", "created_on", "created_by",
                                       "updated_on", "updated_by")
VALUES ('release-track', 'ReleaseTrack', 'ReleaseTrack', ARRAY ['release'], false, now(), 1, now(), 1);




ALTER TABLE roles ADD COLUMN IF NOT EXISTS "release" text;
ALTER TABLE roles ADD COLUMN IF NOT EXISTS "release_track" text;