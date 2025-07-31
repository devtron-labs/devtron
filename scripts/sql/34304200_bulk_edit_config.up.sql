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

-- Insert data from bulk_edit_config for 'v1beta1/application'
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
      "type": "string",
      "const": "batch/v1beta1"
    },
    "kind": {
      "type": "string",
      "enum": ["application", "Application"]
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

-- Insert data from bulk_edit_config for 'v1beta2/application'
INSERT INTO "public"."bulk_edit_config" ("api_version", "kind", "readme", "schema", "created_on", "created_by")
SELECT 'v1beta2', 'application','# New Script (v1beta2)

## Introduction

**v1beta2** is the latest YAML script to perform bulk edits to Deployment Templates, ConfigMaps, or Secrets across multiple applications. This version is currently available only to Enterprise users.

The script provides selectors for choosing the project, application, and environment within which you wish to edit the configs (can be used in combo). Moreover, you now have granular control over the update and delete operations you wish to perform on the configs.

> **RBAC**: User needs to have [permissions](https://docs.devtron.ai/global-configurations/authorization/user-access#grant-specific-permissions) to apps and environments to edit their configs.

### Tree Structure of the Bulk Edit Script

Below is the visual structure of the script. Refer [Examples](#examples-with-full-script) and [YAML Template](#combined-yaml-template) to know more.

```yaml
v1beta2 Script
├── apiVersion                            # (API version is batch/v1beta2)
├── kind                                  # (Resource kind is Application)
└── spec                                  # (Main configuration)
    ├── selectors                         # (Target apps filter)
    │   └── match                         # (Filter logic used in solo or combo)
    │       ├── project                   # (Project filters)
    │       │   ├── includes              # (Projects to include)
    │       │   │   └── names             # (Array of project names)
    │       │   └── excludes              # (Projects to exclude)
    │       │       └── names             # (Array of project names)
    │       │
    │       ├── app                       # (Application filters)
    │       │   ├── includes              # (Apps to include)
    │       │   │   └── names             # (Array of app names)
    │       │   └── excludes              # (Apps to exclude)
    │       │       └── names             # (Array of app names)
    │       │
    │       └── env                       # (Environment filters)
    │           ├── includes              # (Envs to include)
    │           │   └── names             # (Array of env names)
    │           ├── excludes              # (Envs to exclude)
    │           │   └── names             # (Array of env names)
    │           └── type                  # (prod / non-prod)
    │
    ├── deploymentTemplate                # (Edit deployment template)
    │   └── spec                          # (Template spec)
    │       ├── match                     # (Filter for DT)
    │       │   ├── include-base-config   # (true or false)
    │       │   └── chart                 # (Filter for charts)
    │       │       ├── name              # (Enter chart type, e.g. Deployment)
    │       │       ├── custom            # (true if it''s a user-uploaded custom chart)
    │       │       └── version           # (Filter for chart version)
    │       │            ├── value        # (Enter chart version, e.g. 4.20.0)
    │       │            └── operator     # (EQUAL | LESS | GREATER | LESS_EQUAL | GREATER_EQUAL)
    │       └── operation                 # (Edit operation)
    │           ├── action                # (update only)
    │           ├── field                 # (values / version)
    │           ├── patchJson             # (Define if operation.field=values)
    │           └── chartVersion          # (Define if operation.field=version)
    │
    ├── configMap                         # (Edit config maps)
    │   └── spec                          # (ConfigMap spec)
    │       ├── match                     # (Filter for ConfigMaps)
    │       │   ├── include-base-config   # (true or false)
    │       │   ├── includes              # (CMs to include)
    │       │   │   └── names             # (Array of config map names)
    │       │   └── excludes              # (CMs to exclude)
    │       │       └── names             # (Array of config map names)
    │       │
    │       └── operation                 # (Edit operation)
    │           ├── action                # (create/update/delete)
    │           ├── field                 # (data)
    │           ├── patchJson             # (JSON patch for update)
    │           └── value                 # (Key-values for create/delete)
    │
    └── secret                            # (Edit secrets)
        └── spec                          # (Secret spec)
            ├── match                     # (Filter for Secrets)
            │   ├── include-base-config   # (true or false)
            │   ├── includes              # (Secrets to include)
            │   │   └── names             # (Array of secret names)
            │   └── excludes              # (Secrets to exclude)
            │       └── names             # (Array of secret names)
            │
            └── operation                 # (Edit operation)
                ├── action                # (create/update/delete)
                ├── field                 # (data)
                ├── patchJson             # (JSON patch for update)
                └── value                 # (Key-values for create/delete)
```

We recommend you to perform bulk edits in 4 parts:

1. [Use Selector Block](#step-1-use-selector-block)
2. [Choose the Configs](#step-2-choose-the-configs)
3. [Specify the Operation](#step-3-specify-the-operation)
4. [Run the Script](#step-4-run-the-script)

---

## Step 1: Use Selector Block

In Devtron, configs like Deployment Template, ConfigMaps, and Secrets are specified within application. So you need to determine the target applications initially.

In the first part, we will look at the selector script required to filter the applications. Only the applications that match your selector logic will have their Deployment Template, ConfigMap, or Secret available for edits.

```yaml
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      project:
        includes:
            names: ["dev-project", "qa-project"]
        excludes:
            names: ["test-project"]
      app:
        includes:
            names: ["%-dashboard"]
        excludes:
            names: ["demo-%"]
      env:
        includes:
            names: ["staging", "dev"]
        excludes:
            names: ["%qa%"]
        type: non-prod
# Next Steps: Add spec.deploymentTemplate, spec.configMap, and/or spec.secret (check Step 2 and 3)
```

You can use `project`, `app`, `env` selectors with the following lists:
* `includes.names` - A list to specify the ones we need to edit.
* `excludes.names` - A list to specify the ones which are not to be edited.

In **includes** and **excludes**, you can give the names of your projects/apps/environments. Additionally, you may use wildcard patterns (like `%-dashboard%`).



---

## Step 2: Choose the Configs

Here you can filter the Deployment Templates, ConfigMaps, and Secrets you wish to edit (refer the block below).

### Deployment Templates

After you add the selector block, add `deploymentTemplate` object.

```yaml
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors: # Add selector logic (check Step 1)
  deploymentTemplate:
    spec:
      match:
        chart:
          name: "Deployment" ## Name of the deployment chart
          custom: false ## Set as true if using your uploaded custom deployment chart
          version:
            value: "4.20.0"
            operator: EQUAL ## Supports "GREATER", "LESS", "GREATER_EQUAL", "LESS_EQUAL"
      operation: # Add operation object (check Step 3)
```

#### What you can do?
* Add or remove Helm values defined in your chart
* Update the chart version itself
* When editing deployment templates, you may choose whether to apply changes to only environment-specific overrides or also to the base configuration:
    * When `true`, your operations apply to base deployment template shared across environments
    * When `false` or omitted, changes apply only to environment-level deployment templates
* Combine with `configMap` and `secret` objects in the same script

> **Next Step**: [Add operations for your Deployment Template(s)](#on-deployment-templates)

### ConfigMaps

```yaml
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors: # Add selector logic (check Step 1)
  configMap:
    spec:
      match:
        include-base-config: true
        includes:
          names: ["qa-cm-%", "prod-cm-%"]
        excludes:
          names: ["%dev%", "%test%"]
      operation: # Add operation object (check Step 3)
```

#### What you can do?
* Add new keys, e.g., `FEATURE_ENABLE_X: true`
* Update existing keys
* Delete keys or the entire ConfigMap by name
* Include Base Configuration
    * This allows updates to the base-level ConfigMap
    * Environment-level ConfigMaps remain unaffected if this flag is not set
* Combine with `deploymentTemplate` and `secret` objects in the same script.

> **Next Step**: [Add operations for your ConfigMap(s)](#on-configmaps)

### Secrets

```yaml
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors: # Add selector logic (check Step 1)
  secret:
    spec:
      match:
        include-base-config: true
        includes:
          names: ["qa-secret-%", "prod-secret-%"]
        excludes:
          names: ["%dev%", "%test%"]
      operation: # Add operation object (check Step 3)
```

#### What you can do?
* Add or update secret keys, e.g., `API_KEY: efd32tr6tsjbf43765`
* Delete keys or the entire Secret by name
* Same usage pattern as ConfigMaps using `action`, `field: data`, and `value`/`patchJson`.
* Include Base Configuration
    * Enables edits on base-level Secret
    * Use this to update secrets across environments from a single source of truth
* Combine with `deploymentTemplate` and `configMap` objects in the same script

> **Next Step**: [Add operations for your Secret(s)](#on-secrets)
---

## Step 3: Specify the Operation

Add the operation to be performed on the selected Deployment Templates, ConfigMaps, and Secrets.


### On Deployment Templates

Supports only using `action: update`, `field: values` or `field: version`, and corresponding `patchJson` or `chartVersion`.

#### Configure memory to `250Mi`

```yaml
...
...
... # Add the following in deploymentTemplate.spec
      operation:
        action: update
        field: values
        patchJson: ''[{ "op": "replace", "path": "/resources/requests/memory", "value": "250Mi" }]''
```

#### Add a new value `ENABLE_AUTOSCALING: true`

```yaml
...
...
... # Add the following in deploymentTemplate.spec
      operation:
        action: update
        field: values
        patchJson: ''[{ "op": "add", "path": "/ENABLE_AUTOSCALING", "value": true }]''
```

#### Update deployment chart version to `4.30.1`

```yaml
...
...
... # Add the following in deploymentTemplate.spec
      operation:
        action: update
        field: version
        chartVersion: "4.30.1"
```

### On ConfigMaps

#### Add `FEATURE_ENABLE_X` key in existing ConfigMap

```yaml
...
...
... # Add the following in configMap.spec
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "add", "path": "/FEATURE_ENABLE_X", "value": "true" }]''
```

#### Update existing key `FEATURE_ENABLE_X`

```yaml
...
...
... # Add the following in configMap.spec
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "replace", "path": "/FEATURE_ENABLE_X", "value": "false" }]''
```

#### Remove existing key `FEATURE_ENABLE_X: true` from ConfigMap

```yaml
...
...
... # Add the following in configMap.spec
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "remove", "path": "/FEATURE_ENABLE_X" }]''
```

#### Delete ConfigMap

```yaml
...
...
... # Add the following in configMap.spec
      operation:
        action: delete
        field: data
        value: banking-cm # In ''value'', enter the name of the ConfigMap to delete
```


### On Secrets

#### Add `API_TOKEN` key in existing secret

```yaml
...
...
... # Add the following in secret.spec
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "add", "path": "/API_TOKEN", "value": "u4hg847598fc" }]''
```

#### Update `DB_PASSWORD` key in secret

```yaml
...
...
... # Add the following in secret.spec
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "replace", "path": "/DB_PASSWORD", "value": "root@123" }]''
```

#### Remove `API_KEY` key in secret

```yaml
...
...
... # Add the following in secret.spec
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "remove", "path": "/API_KEY" }]''
```

#### Add multiple secret keys

```yaml
...
...
... # Add the following in secret.spec
      operation:
        action: update
        field: data
        patchJson: ''[
          { "op": "add", "path": "/TOKEN", "value": "shc24235" },
          { "op": "replace", "path": "/API_KEY", "value": "u4hg847598fc" }
        ]''
```

#### Delete Secret

```yaml
...
...
... # Add the following in secret.spec
      operation:
        action: delete
        field: data
        value: banking-secret # In ''value'', enter the name of the secret to delete
```

---

## Step 4: Run the Script

Before running the script, make sure to check the impacted applications and configs, by clicking the **Show Impacted Objects** button. We recommend you to do this just so you don''t end up unintentionally editing any config (Deployment Templates, ConfigMaps, and Secrets).

Next, click **Run** to execute the script. The output of the script execution will be shown in the **Output** tab in the bottom drawer.

---

## Examples (With Full Script)

### Edit Deployment Template

**CASE 1**: Update `replicaCount` in only Base Deployment Templates of `devtron` project

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      project:
        includes:
          names:
            - devtron
  deploymentTemplate:
    spec:
      match:
        include-base-config: true
      operation:
        action: update
        field: values
        patchJson: ''[{ "op": "replace", "path": "/replicaCount", "value": 2 }]''
```

**CASE 2**: Remove `replicaCount` in only Base Deployment Templates of all applications (names) that end with `-sanity-app`

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      app:
        includes:
          names:
            - "%-sanity-app"
  deploymentTemplate:
    spec:
      match:
        include-base-config: true
      operation:
        action: update
        field: values
        patchJson: ''[{ "op": "remove", "path": "/replicaCount" }]''
```

**CASE 3**: Add `replicaCount` in both (Base + Env Override) Deployment Templates of all applications (names) having at least one prod environment (excluding any environment name containing `virtual`)

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      env:
        type: prod
        excludes:
          names:
            - "%virtual%"
  deploymentTemplate:
    spec:
      match:
        include-base-config: true
      operation:
        action: update
        field: values
        patchJson: ''[{ "op": "add", "path": "/replicaCount", "value": 2 }]''
```

**CASE 4**: Change Deployment Chart Version (Only Env Override) to `4.30.0` for all `non-prod` environments having `Rollout Deployment` chart and version `<= 4.20.0`

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      env:
        type: non-prod
  deploymentTemplate:
    spec:
      match:
        chart:
          name: "Rollout Deployment"
          version:
            value: 4.20.0
            operator: LESS_EQUAL
      operation:
        action: update
        field: version
        chartVersion: 4.30.0
```

### Edit ConfigMap

**CASE 1**: Update `USE_GIT_CLI` in `orchestrator-cm` in only Base ConfigMap of all projects (names) that starts with `go-lang`

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      project:
        includes:
          names:
            - "go-lang%"
  configMap:
    spec:
      match:
        include-base-config: true
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "replace", "path": "/USE_GIT_CLI", "value": "true" }]''
```

**CASE 2**: Remove `USE_GIT_CLI` in `orchestrator-cm` in only Base ConfigMap of all applications, except for the ConfigMaps with name `orchestrator-cm`

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      app:
        includes:
          names:
            - "%%" # %% Wildcard for all applications
  configMap:
    spec:
      match:
        include-base-config: true
        excludes:
          names:
            - orchestrator-cm
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "remove", "path": "/USE_GIT_CLI" }]''
```

**CASE 3**: Add `USE_GIT_CLI` in `orchestrator-cm` in both (Base + Env Override) ConfigMaps of all applications (names) having at least one `non-prod` environment (excluding any environment name containing `virtual`)

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      env:
        type: non-prod
        excludes:
          names:
            - "%virtual%"
  configMap:
    spec:
      match:
        include-base-config: true
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "add", "path": "/USE_GIT_CLI", "value": "true" }]''
```

**CASE 4**: Delete a ConfigMap named `orchestrator-cm` (Only From Env Override) from all `non-prod` environments except for the applications with name `orchestrator-app`

```YAML
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      app:
        excludes:
          names:
            - orchestrator-app
      env:
        type: non-prod
  configMap:
    spec:
      operation:
        action: delete
        value: orchestrator-cm
```

---

## Combined YAML Template

You can use the below script as a template if you wish to edit Deployment Templates, ConfigMaps, Secrets of one or more apps in bulk.

```yaml
apiVersion: batch/v1beta2
kind: Application
spec:
  selectors:
    match:
      project:
        includes:
          names: ["dev"]
        excludes:
          names: ["test"]
      app:
        includes:
          names: ["%-dashboard", "%-server"]
        excludes:
          names: ["%demo-%", "%test-%"]
      env:
        type: non-prod
  deploymentTemplate:
    spec:
      match:
        include-base-config: true
        chart:
          name: "Deployment"
          custom: false
          version:
            value: "4.20.0"
            operator: LESS_EQUAL
  configMap:
    spec:
      match:
        include-base-config: true
        includes:
          names: ["qa-cm-%", "prod-cm-%"]
        excludes:
          names: ["%dev%", "%test%"]
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "add", "path": "/FEAT_TEST_ENABLE", "value": "true" },{"op": "replace","path":"/LOG_LEVEL","value": "-1"}]''
  secret:
    spec:
      match:
        include-base-config: true
        includes:
          names: ["qa-secret-%", "prod-secret-%"]
        excludes:
          names: ["%dev%", "%test%"]
      operation:
        action: update
        field: data
        patchJson: ''[{ "op": "add", "path": "/DB_PASSWORD", "value": "********" },{"op": "replace","path":"/ADMIN_PASSWORD","value": "********"}]''
```
','{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "Bulk Edit Application batch/v1beta2 Schema",
  "type": "object",
  "description": "Input Script for bulk edit",
  "required": ["apiVersion", "kind", "spec"],
  "properties": {
    "apiVersion": {
      "type": "string",
      "enum": ["batch/v1beta2"],
      "description": "API version of the bulk edit schema"
    },
    "kind": {
      "type": "string",
      "enum": ["application", "Application"],
      "description": "Resource kind, must be ''application'' or ''Application''"
    },
    "spec": {
      "type": "object",
      "required": ["selectors"],
      "description": "Specification for the bulk edit operation",
      "properties": {
        "selectors": {
          "$ref": "#/definitions/Selectors",
          "description": "Criteria to select target applications, projects, or environments"
        },
        "deploymentTemplate": {
          "$ref": "#/definitions/DeploymentTemplate",
          "description": "Deployment template update specification"
        },
        "configMap": {
          "$ref": "#/definitions/ConfigMap",
          "description": "ConfigMap update specification"
        },
        "secret": {
          "$ref": "#/definitions/Secret",
          "description": "Secret update specification"
        }
      },
      "anyOf": [
        { "required": ["deploymentTemplate"] },
        { "required": ["configMap"] },
        { "required": ["secret"] }
      ],
      "additionalProperties": false
    }
  },
  "additionalProperties": false,
  "definitions": {
    "Selectors": {
      "type": "object",
      "description": "Selectors to filter projects, apps, or environments",
      "properties": {
        "match": {
          "type": "object",
          "description": "Match criteria for selection",
          "properties": {
            "project": {
              "$ref": "#/definitions/ProjectSelector",
              "description": "Project selection criteria"
            },
            "app": {
              "$ref": "#/definitions/AppSelector",
              "description": "Application selection criteria"
            },
            "env": {
              "$ref": "#/definitions/EnvSelector",
              "description": "Environment selection criteria"
            }
          },
          "anyOf": [
            { "required": ["project"] },
            { "required": ["app"] },
            { "required": ["env"] }
          ]
        }
      },
      "additionalProperties": false
    },
    "ProjectSelector": {
      "type": "object",
      "description": "Selector for projects",
      "properties": {
        "includes": {
          "$ref": "#/definitions/NameIncludesExcludes",
          "description": "Projects to include"
        },
        "excludes": {
          "$ref": "#/definitions/NameIncludesExcludes",
          "description": "Projects to exclude"
        }
      },
      "anyOf": [
        { "required": ["includes"] },
        { "required": ["excludes"] }
      ]
    },
    "AppSelector": {
      "type": "object",
      "description": "Selector for applications",
      "properties": {
        "includes": {
          "$ref": "#/definitions/NameIncludesExcludes",
          "description": "Applications to include"
        },
        "excludes": {
          "$ref": "#/definitions/NameIncludesExcludes",
          "description": "Applications to exclude"
        }
      },
      "anyOf": [
        { "required": ["includes"] },
        { "required": ["excludes"] }
      ]
    },
    "EnvSelector": {
      "type": "object",
      "description": "Selector for environments",
      "properties": {
        "type": {
          "type": "string",
          "enum": ["prod", "non-prod"],
          "description": "Type of environment: production or non-production"
        },
        "includes": {
          "$ref": "#/definitions/NameIncludesExcludes",
          "description": "Environments to include"
        },
        "excludes": {
          "$ref": "#/definitions/NameIncludesExcludes",
          "description": "Environments to exclude"
        }
      },
      "anyOf": [
        { "required": ["type"] },
        { "required": ["includes"] },
        { "required": ["excludes"] }
      ]
    },
    "DeploymentTemplate": {
      "type": "object",
      "required": ["spec"],
      "description": "Deployment template update specification",
      "properties": {
        "spec": {
          "type": "object",
          "required": ["operation"],
          "description": "Specification for deployment template operation",
          "properties": {
            "match": {
              "type": "object",
              "description": "Criteria to match deployment templates",
              "properties": {
                "include-base-config": {
                  "type": "boolean",
                  "description": "Whether to include base configuration"
                },
                "chart": {
                  "type": "object",
                  "description": "Chart selection criteria",
                  "properties": {
                    "name": {
                      "type": "string",
                      "description": "Deployment chart name"
                    },
                    "custom": {
                      "type": "boolean",
                      "description": "Whether the deployment chart is user uploaded (Not devtron managed)"
                    },
                    "version": {
                      "type": "object",
                      "required": ["value", "operator"],
                      "description": "Chart version criteria",
                      "properties": {
                        "value": {
                          "type": "string",
                          "pattern": "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$",
                          "description": "Chart version value (semver)"
                        },
                        "operator": {
                          "type": "string",
                          "enum": ["EQUAL", "GREATER", "LESS", "GREATER_EQUAL", "LESS_EQUAL"],
                          "description": "Operator for version comparison"
                        }
                      }
                    }
                  }
                }
              }
            },
            "operation": {
              "type": "object",
              "required": ["action", "field"],
              "description": "Operation to perform on deployment template",
              "properties": {
                "action": {
                  "type": "string",
                  "enum": ["update"],
                  "description": "Action to perform (only ''update'' supported)"
                },
                "field": {
                  "type": "string",
                  "enum": ["values", "version"],
                  "description": "Field to update (values or version). For values, use patchJson; for version, use chartVersion"
                },
                "patchJson": { "type": "string", "description": "JSON patch to apply" },
                "chartVersion": {
                  "type": "string",
                  "pattern": "^(0|[1-9]\\d*)\\.(0|[1-9]\\d*)\\.(0|[1-9]\\d*)(?:-((?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9]\\d*|\\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+([0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$",
                  "description": "Chart version to update to"
                }
              }
            }
          }
        }
      }
    },
    "ConfigMap": {
      "type": "object",
      "required": ["spec"],
      "description": "ConfigMap update specification",
      "properties": {
        "spec": {
          "type": "object",
          "required": ["operation"],
          "description": "Specification for ConfigMap operation",
          "properties": {
            "match": {
              "type": "object",
              "description": "Criteria to match ConfigMaps",
              "properties": {
                "include-base-config": {
                  "type": "boolean",
                  "description": "Whether to include base configuration"
                },
                "includes": {
                  "$ref": "#/definitions/NameIncludesExcludes",
                  "description": "ConfigMaps to include"
                },
                "excludes": {
                  "$ref": "#/definitions/NameIncludesExcludes",
                  "description": "ConfigMaps to exclude"
                }
              }
            },
            "operation": {
              "type": "object",
              "required": ["action"],
              "description": "Operation to perform on ConfigMap",
              "properties": {
                "action": {
                  "type": "string",
                  "enum": ["create", "update", "delete"],
                  "description": "Action to perform (create, update, delete). For update, use field and patchJson and for create/ delete, use value."
                },
                "field": {
                  "type": "string",
                  "enum": ["data"],
                  "description": "Field to update. Only ''data'' is supported. Use patchJson for updates."
                },
                "patchJson": {
                  "type": "string",
                  "description": "JSON patch to apply"
                },
                "value": {
                  "type": "string",
                  "description": "Value to set. For action ''create'', this should be a JSON string representing the ConfigMap data and for ''delete'', it should be the name of the ConfigMap to delete."
                }
              }
            }
          }
        }
      }
    },
    "Secret": {
      "type": "object",
      "required": ["spec"],
      "description": "Secret update specification",
      "properties": {
        "spec": {
          "type": "object",
          "required": ["operation"],
          "description": "Specification for Secret operation",
          "properties": {
            "match": {
              "type": "object",
              "description": "Criteria to match Secrets",
              "properties": {
                "include-base-config": {
                  "type": "boolean",
                  "description": "Whether to include base configuration"
                },
                "includes": {
                  "$ref": "#/definitions/NameIncludesExcludes",
                  "description": "Secrets to include"
                },
                "excludes": {
                  "$ref": "#/definitions/NameIncludesExcludes",
                  "description": "Secrets to exclude"
                }
              }
            },
            "operation": {
              "type": "object",
              "required": ["action"],
              "description": "Operation to perform on Secret",
              "properties": {
                "action": {
                  "type": "string",
                  "enum": ["create", "update", "delete"],
                  "description": "Action to perform (create, update, delete). For update, use field and patchJson; for create/delete, use value."
                },
                "field": {
                  "type": "string",
                  "enum": ["data"],
                  "description": "Field to update. Only ''data'' is supported. Use patchJson for updates."
                },
                "patchJson": {
                  "type": "string",
                  "description": "JSON patch to apply"
                },
                "value": {
                  "type": "string",
                  "description": "Value to set. For action ''create'', this should be a JSON string representing the Secret data and for ''delete'', it should be the name of the Secret to delete."
                }
              }
            }
          }
        }
      }
    },
    "NameIncludesExcludes": {
      "type": "object",
      "required": ["names"],
      "description": "Names to include or exclude",
      "properties": {
        "names": {
          "type": "array",
          "items": {
            "type": "string",
            "pattern": "^[a-z%]+[a-z0-9%\\-\\?]*[a-z0-9%]+$",
            "description": "Name pattern"
          },
          "minItems": 1,
          "description": "List of names"
        }
      },
      "additionalProperties": false
    }
  }
}', NOW(), 1
WHERE NOT EXISTS (
    SELECT 1 FROM "public"."bulk_edit_config"
    WHERE api_version = 'v1beta2' AND kind = 'application'
);

COMMIT;
