DROP TABLE IF EXISTS resource_qualifier_mapping_criteria CASCADE;

DELETE from devtron_resource_searchable_key ds where ds."name" in ('GLOBAL_ID', 'BASE_DEPLOYMENT_TEMPLATE');
