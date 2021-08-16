UPDATE "public"."bulk_update_readme"
SET "script" = '{"kind": "Application", "spec": {"envIds": [1, 2, 3], "global": false, "excludes": {"names": ["%xyz%"]}, "includes": {"names": ["%abc%"]}, "deploymentTemplate": {"spec": {"patchJson": "Enter Patch String"}}, "configMap": {"spec": { "names": ["abc"],"patchJson": "Enter Patch String"}},"secret": {"spec": { "names": ["abc"],"patchJson": "Enter Patch String"}}}, "apiVersion": "core/v1beta1"}',
    "readme" = '# Bulk Update - Application

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


The following tables list the configurable parameters of the Payload component in the Script and their description along with example.

| Parameter                      | Description                        | Example                                                    |
| -------------------------- | ---------------------------------- | ---------------------------------------------------------- |
|`includes.names `        | Will filter apps having exact string or similar substrings                 | `["app%","%abc", "xyz"]` (will include all apps having `"app%"` **OR** `"%abc"` as one of their substring, example - app1, app-test, test-abc etc. **OR** application with name xyz)    |
| `excludes.names`          | Will filter apps not having exact string or similar substrings.              | `["%z","%y", "abc"]`       (will filter out all apps having `"%z"` **OR** `"%y"` as one of their substring, example - appz, test-app-y etc. **OR** application with name abc)                                        |
| `envIds`       | List of envIds to be updated for the selected applications.           | `[1,2,3]`                                                   |
| `global`       | Flag to update global deployment template of applications.            | `true`,`false`                                                        |
| `deploymentTemplate.spec.patchJson`       | String having the update operation(you can apply more than one changes at a time). It supports [JSON patch ](http://jsonpatch.com/) specifications for update. | `''[ { "op": "add", "path": "/MaxSurge", "value": 1 }, { "op": "replace", "path": "/GracePeriod", "value": "30" }]''` |
| `configMap.spec.names`      | Names of all ConfigMaps to be updated. | `configmap1`,`configmap2`,`configmap3` |
| `secret.spec.names`      | Names of all Secrets to be updated. | `secret1`,`secret2`|
| `configMap.spec.patchJson` / `secret.spec.patchJson`       | String having the update operation for ConfigMaps/Secrets(you can apply more than one changes at a time). It supports [JSON patch ](http://jsonpatch.com/) specifications for update. | `''[{ "op": "add", "path": "/{key}", "value": "{value}" },{"op": "replace","path":"/{key}","value": "{value}"}]''`(Replace the `{key}` part to the key you want to perform operation on & `{value}`is the key''s corresponding value |
' WHERE "id" = 1;