DROP TABLE "public"."devtron_resource_task_run";

DROP SEQUENCE IF EXISTS id_devtron_resource_task_run;

ALTER TABLE devtron_resource_object
    DROP COLUMN identifier;

ALTER TABLE devtron_resource_object_audit
    DROP COLUMN audit_operation_path;

DELETE FROM devtron_resource_schema where devtron_resource_id in (select id from devtron_resource where kind in('release-track', 'release'));

DELETE FROM devtron_resource where kind in('release-track', 'release');

DELETE FROM global_policy where policy_of in('RELEASE_STATUS', 'RELEASE_ACTION_CHECK');