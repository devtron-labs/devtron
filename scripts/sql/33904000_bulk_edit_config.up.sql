BEGIN;

-- Create a new sequence
CREATE SEQUENCE IF NOT EXISTS id_seq_bulk_edit_config;

-- Create a new table
CREATE TABLE IF NOT EXISTS "public"."bulk_edit_config" (
    "id"            INTEGER             NOT NULL DEFAULT nextval('id_seq_workflow_config_snapshot'::regclass),
    "api_version"   VARCHAR(255) NOT NULL,
    "kind"          VARCHAR(255) NOT NULL,
    "readme"        TEXT,
    "schema"        TEXT,
    "created_on"    timestamptz NOT NULL,
    "created_by"    INTEGER NOT NULL,
    "updated_on"    timestamptz,
    "updated_by"    INTEGER,
    PRIMARY KEY ("id"),
    UNIQUE ("api_version", "kind")
);

-- Insert data from bulk_update_readme (assuming only one row for 'v1beta1/application')
INSERT INTO "public"."bulk_edit_config" ("api_version", "kind", "readme", "schema", "created_on", "created_by")
SELECT 'v1beta1', 'application','# Bulk Update - Application

This feature helps you to update Deployment Template, ConfigMaps & Secrets for multiple apps in one go! You can filter the apps on the basis of environments, global flag, and app names(we provide support for both substrings included and excluded in the app name).

## Example

Example below will select all applications having `abc and xyz` present in their name and out of those will exclude applications having `abcd and xyza` in their name. Since global flag is false and envId 23 is provided, it will make changes in envId 23 and not in global deployment template for this application.

If you want to update globally then please set `global: true`. If you have provided envId by deployment template, configMap or secret is not overridden for that particular environment then it will not apply the changes.
Also, of all the provided names of configMaps/secrets, for every app & environment override only the name that are present in them will be considered.

```
apiVersion: batch/v1beta1
kind: Application
spec:
  includes:
    names:
    - "%abc%"
    - "%xyz%"
  excludes:
    names:
    - "%abcd%"
    - "%xyza%"
  envIds:
  - 23
  global: false
  deploymentTemplate:
    spec:
      patchJson: ''[{ "op": "add", "path": "/MaxSurge", "value": 1 },{"op": "replace","path":"/GracePeriod","value": "30"}]''
  configMap:
    spec:
      names:
      - "configmap1"
      - "configmap2"
      - "configmap3"
      patchJson: ''[{ "op": "add", "path": "/{key}", "value": "{value}" },{"op": "replace","path":"/{key}","value": "{value}"}]''
  secret:
    spec:
      names:
      - "secret1"
      - "secret2"
      patchJson: ''[{ "op": "add", "path": "/{key}", "value": "{value}" },{"op": "replace","path":"/{key}","value": "{value}"}]''
```

## Payload Configuration


The following tables list the configurable parameters of the Payload component in the Script and their description along with example. Also, if you do not need to apply updates on all the tasks, i.e. Deployment Template, ConfigMaps & Secrets, leave the Spec object empty for that respective task.

| Parameter                      | Description                        | Example                                                    |
| -------------------------- | ---------------------------------- | ---------------------------------------------------------- |
|`includes.names `        | Will filter apps having exact string or similar substrings                 | `["app%","%abc", "xyz"]` (will include all apps having `"app%"` **OR** `"%abc"` as one of their substring, example - app1, app-test, test-abc etc. **OR** application with name xyz)    |
| `excludes.names`          | Will filter apps not having exact string or similar substrings.              | `["%z","%y", "abc"]`       (will filter out all apps having `"%z"` **OR** `"%y"` as one of their substring, example - appz, test-app-y etc. **OR** application with name abc)                                        |
| `envIds`       | List of envIds to be updated for the selected applications.           | `[1,2,3]`                                                   |
| `global`       | Flag to update global deployment template of applications.            | `true`,`false`                                                        |
| `deploymentTemplate.spec.patchJson`       | String having the update operation(you can apply more than one changes at a time). It supports [JSON patch ](http://jsonpatch.com/) specifications for update. | `''[ { "op": "add", "path": "/MaxSurge", "value": 1 }, { "op": "replace", "path": "/GracePeriod", "value": "30" }]''` |
| `configMap.spec.names`      | Names of all ConfigMaps to be updated. | `configmap1`,`configmap2`,`configmap3` |
| `secret.spec.names`      | Names of all Secrets to be updated. | `secret1`,`secret2`|
| `configMap.spec.patchJson` / `secret.spec.patchJson`       | String having the update operation for ConfigMaps/Secrets(you can apply more than one changes at a time). It supports [JSON patch ](http://jsonpatch.com/) specifications for update. | `''[{ "op": "add", "path": "/{key}", "value": "{value}" },{"op": "replace","path":"/{key}","value": "{value}"}]''`(Replace the `{key}` part to the key you want to perform operation on & the `{value}`is the key''s corresponding value |
','{
  "$schema": "https://json-schema.org/draft/2020-12/schema",
  "title": "Bulk Edit Application batch/v1beta1 Schema",
  "type": "object",
  "description": "Input Script for bulk edit",
  "required": [
    "apiVersion",
    "kind",
    "spec"
  ],
  "properties": {
    "apiVersion": {
      "const": "v1beta1"
    },
    "kind": {
      "const": "Application"
    },
    "spec": {
      "type": "object",
      "anyOf": [
        {
          "required": [
            "deploymentTemplate"
          ]
        },
        {
          "required": [
            "configMap"
          ]
        },
        {
          "required": [
            "secret"
          ]
        }
      ],
      "properties": {
        "includes": {
          "type": "object",
          "properties": {
            "names": {
              "type": "array",
              "items": {
                "type": "string",
                "pattern": "^[a-z%]+[a-z0-9%\\-\\?]*[a-z0-9%]+$"
              },
              "description": "Array of application names to be included"
            }
          }
        },
        "excludes": {
          "type": "object",
          "properties": {
            "names": {
              "type": "array",
              "items": {
                "type": "string",
                "pattern": "^[a-z%]+[a-z0-9%\\-\\?]*[a-z0-9%]+$"
              },
              "description": "Array of application names to be excluded"
            }
          }
        },
        "envIds": {
          "type": "array",
          "items": {
            "type": "integer",
            "exclusiveMinimum": 0
          },
          "description": "Array of Environment Ids of dependent apps"
        },
        "global": {
          "type": "boolean",
          "description": "Flag for updating base Configurations of dependent apps"
        },
        "deploymentTemplate": {
          "type": "object",
          "properties": {
            "spec": {
              "type": "object",
              "properties": {
                "patchData": {
                  "type": "string",
                  "description": "String with details of the patch to be used for updating"
                }
              }
            }
          }
        },
        "configMap": {
          "type": "object",
          "properties": {
            "names": {
              "type": "array",
              "items": {
                "type": "string",
                "pattern": "^[a-z]+[a-z0-9\\-\\?]*[a-z0-9]+$"
              },
              "description": "Name of all ConfigMaps to be updated"
            },
            "patchData": {
              "type": "string",
              "description": "String with details of the patch to be used for updating"
            }
          }
        },
        "secret": {
          "type": "object",
          "properties": {
            "names": {
              "type": "array",
              "items": {
                "type": "string"
              },
              "description": "Name of all Secrets to be updated"
            },
            "patchData": {
              "type": "string",
              "description": "String with details of the patch to be used for updating"
            }
          }
        }
      }
    }
  }
}', NOW(), 1
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."bulk_edit_config"
    WHERE api_version = 'v1beta1' AND kind = 'application'
);

COMMIT;
