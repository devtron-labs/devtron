ALTER TABLE devtron_resource_object_dep_relations ADD COLUMN IF NOT EXISTS dependency_object_identifier VARCHAR(150);
ALTER TABLE devtron_resource_object_dep_relations ADD COLUMN IF NOT EXISTS component_object_identifier VARCHAR(150);
