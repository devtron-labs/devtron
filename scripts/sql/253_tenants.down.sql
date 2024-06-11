DROP TABLE IF EXISTS "public"."devtron_resource_object_dep_mapping";


DELETE FROM devtron_resource_schema where devtron_resource_id in (select id from devtron_resource where kind in('tenant', 'installation'));

DELETE FROM devtron_resource where kind in('tenant', 'installation');