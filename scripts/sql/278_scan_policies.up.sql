
-- severity 3 is for high and 5 is for unknown
INSERT INTO "public"."cve_policy_control" ("global", "cluster_id", "env_id", "app_id", "cve_store_id", "action", "severity", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
                                                     ('t', NULL, NULL, NULL, NULL, '1', '3', 'f', 'now()', '1', 'now()', '1'),
                                                     ('t', NULL, NULL, NULL, NULL, '1', '5', 'f', 'now()', '1', 'now()', '1');

