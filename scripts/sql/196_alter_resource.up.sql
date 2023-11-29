INSERT INTO devtron_resource(kind, display_name, icon, parent_kind_id, deleted, created_on, created_by, updated_on,
                             updated_by)
VALUES ('cd-pipeline', 'Cd Pipeline', '', 0, false, now(), 1, now(), 1);


INSERT INTO devtron_resource_schema(devtron_resource_id, version, schema, latest, created_on, created_by, updated_on,
                                    updated_by)
VALUES ((select id from devtron_resource where kind = 'cd-pipeline'), 'v1',
        '{
          "$schema": "https://json-schema.org/draft/2020-12/schema",
          "title": "CD Pipeline Schema",
          "type": "object",
          "properties": {
            "version": {
              "type": "string"
            },
            "kind": {
              "type": "string"
            },
            "overview": {
              "type": "object"
            },
            "actions": {
              "type": "object"
            },
            "dependencies": {
              "type": "array"
            }
          },
          "required": [
            "version",
            "kind"
          ]
        }',
        true, now(), 1, now(), 1);