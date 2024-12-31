alter table devtron_resource
    drop column is_exposed;

alter table devtron_resource
    drop column description;

alter table devtron_resource_schema
    drop column sample_schema;

DROP TABLE "public"."devtron_resource_schema_audit";