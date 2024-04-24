ALTER TABLE devtron_resource_object
    DROP COLUMN identifier;

ALTER TABLE devtron_resource_object_audit
    DROP COLUMN audit_operation_path;