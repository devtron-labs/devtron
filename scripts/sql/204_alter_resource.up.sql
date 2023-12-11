alter table devtron_resource
    add column description text;

alter table devtron_resource
    add column is_exposed bool not null default true;

update devtron_resource
set is_exposed= false
where kind = 'application';

update devtron_resource_schema
set is_exposed= false
where kind = 'cd-pipeline';


alter table devtron_resource_schema
    add column sample_schema json;
