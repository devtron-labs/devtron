-- Sequence and defined type
CREATE SEQUENCE IF NOT EXISTS id_seq_bulk_update_readme;

-- Table Definition
CREATE TABLE "public"."bulk_update_readme" (
                                               "id" int4 NOT NULL DEFAULT nextval('id_seq_bulk_update_readme'::regclass),
                                               "operation" varchar(255) NOT NULL,
                                               "readme" text,
                                               "script" jsonb,
                                               PRIMARY KEY ("id")
);

INSERT INTO "public"."bulk_update_readme" ("id", "operation", "readme", "script") VALUES
(1, 'v1beta1/application', '# Bulk Update - Deployment Template

This feature helps you to update deployment template for multiple apps in one go! You can filter the apps on the basis of environments, global flag, and app names(we provide support for both substrings included and excluded in the app name).

## Script

This is the piece of code which works as the input and has to be pasted in the code editor for achieving bulk updation task.

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
    - "%abc%"
    - "%xyz%"
  envIds: []
  global: true
  deploymentTemplate:
    spec:
      patchJson: Enter Patch String
```

## Payload Configuration


The following tables list the configurable parameters of the Payload component in the Script and their description along with example.

| Parameter                      | Description                        | Example                                                    |
| -------------------------- | ---------------------------------- | ---------------------------------------------------------- |
|`includes.names `        | Will filter apps having similar substrings                 | `["app%","%abc"]` (will include all apps having `"app%"` **OR** `"%abc"` as one of their substring, example - app1, app-test, test-abc etc.)    |
| `excludes.names`          | Will filter apps not having similar substrings              | `["%z","%y"]`       (will filter out all apps having `"%z"` **OR** `"%y"` as one of their substring, example - appz, test-app-y etc.)                                        |
| `envIds`       |Will filter apps by all environment with IDs in this array              | `[1,2,3]`                                                   |
| `global`       | Will filter apps by global flag            | `true`,`false`                                                        |
| `patchJson`      | String having the update operation(you can apply more than one changes at a time) | `''[ { "op": "add", "path": "/MaxSurge", "value": 1 }, { "op": "replace", "path": "/GracePeriod", "value": "30" }]''` |

Note - We use [JSON patch](http://jsonpatch.com/) logic for updation, visit the link for more info on this.', '{"kind": "Application", "spec": {"envIds": [1, 2, 3], "global": false, "excludes": {"names": ["%xyz%"]}, "includes": {"names": ["%abc%"]}, "deploymentTemplate": {"spec": {"patchJson": "Enter Patch String"}}}, "apiVersion": "core/v1beta1"}');