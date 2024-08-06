
-- severity 3 is for high and 5 is for unknown
INSERT INTO "public"."cve_policy_control" ("global", "cluster_id", "env_id", "app_id", "cve_store_id", "action", "severity", "deleted", "created_on", "created_by", "updated_on", "updated_by") VALUES
                                                                                                                                                                                                    ('t', NULL, NULL, NULL, NULL, '1', '3', 'f', 'now()', '1', 'now()', '1'),
                                                                                                                                                                                                    ('t', NULL, NULL, NULL, NULL, '1', '5', 'f', 'now()', '1', 'now()', '1');

-- UPDATE THE NULL VALUES IN THE STANDARD_SEVERITY COLUMN WITH THE SEVERITY COLUMN
-- standard_severity column was introduced in 148_trivy_image_scanning.up.sql. all the scans happened before that will contains null values in standard_severity column.
UPDATE "public"."cve_store" SET standard_severity = severity WHERE standard_severity IS NULL;