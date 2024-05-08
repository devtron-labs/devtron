-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_bulk_update_readme;

-- Table Definition
CREATE TABLE "public"."bulk_update_readme" (
                                               "id" int4 NOT NULL DEFAULT nextval('id_seq_bulk_update_readme'::regclass),
                                               "resource" varchar(255) NOT NULL,
                                               "readme" text,
                                               "script" jsonb,
                                               PRIMARY KEY ("id")
);

INSERT INTO "public"."bulk_update_readme" ("id", "resource", "readme", "script") VALUES
    (1, 'v1beta1/application', '# Bulk Update - Application
This feature helps you to update deployment template for multiple apps in one go! You can filter the apps on the basis of environments, global flag, and app names(we provide support for both substrings included and excluded in the app name).
## Example
Example below will select all applications having `abc and xyz` present in their name and out of those will exclude applications having `abcd and xyza` in their name. Since global flag is false and envId 23 is provided, it will make changes in envId 23 and not in global deployment template for this application.
If you want to update global deployment template then please set `global: true`.  If you have provided envId by deployment template is not overridden for that particular environment then it will not apply the changes.
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
```
## Payload Configuration
The following tables list the configurable parameters of the Payload component in the Script and their description along with example.
| Parameter                      | Description                        | Example                                                    |
| -------------------------- | ---------------------------------- | ---------------------------------------------------------- |
|`includes.names `        | Will filter apps having exact string or similar substrings                 | `["app%","%abc", "xyz"]` (will include all apps having `"app%"` **OR** `"%abc"` as one of their substring, example - app1, app-test, test-abc etc. **OR** application with name xyz)    |
| `excludes.names`          | Will filter apps not having exact string or similar substrings.              | `["%z","%y", "abc"]`       (will filter out all apps having `"%z"` **OR** `"%y"` as one of their substring, example - appz, test-app-y etc. **OR** application with name abc)                                        |
| `envIds`       | List of envIds to be updated for the selected applications           | `[1,2,3]`                                                   |
| `global`       | Flag to update global deployment template of applications            | `true`,`false`                                                        |
| `patchJson`      | String having the update operation(you can apply more than one changes at a time). It supports [JSON patch ](http://jsonpatch.com/) specifications for update. | `''[ { "op": "add", "path": "/MaxSurge", "value": 1 }, { "op": "replace", "path": "/GracePeriod", "value": "30" }]''` |
', '{"kind": "Application", "spec": {"envIds": [1, 2, 3], "global": false, "excludes": {"names": ["%xyz%"]}, "includes": {"names": ["%abc%"]}, "deploymentTemplate": {"spec": {"patchJson": "Enter Patch String"}}}, "apiVersion": "core/v1beta1"}');

CREATE INDEX ON bulk_update_readme (resource);
