DROP TABLE IF EXISTS "public"."resource_filter";

DROP SEQUENCE IF EXISTS public.resource_filter_seq;

ALTER TABLE resource_qualifier_mapping DROP COLUMN IF EXISTS resource_type;

ALTER TABLE resource_qualifier_mapping RENAME TO variable_scope;

ALTER TABLE variable_scope RENAME COLUMN resource_id TO variable_definition_id;

DELETE from devtron_resource_searchable_key where name = 'PROJECT_ID';

ALTER TABLE variable_scope ADD CONSTRAINT variable_scope_variable_id_fkey FOREIGN KEY ("variable_definition_id") REFERENCES "public"."variable_definition" ("id");

ALTER TABLE variable_data ADD CONSTRAINT variable_data_variable_id_fkey FOREIGN KEY ("variable_scope_id") REFERENCES "public"."variable_scope" ("id");



