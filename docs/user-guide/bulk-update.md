# Bulk Updates
This feature helps you to update deployment template for multiple apps in one go! You can filter the apps on the basis of environments, global flag, and app names(we provide support for both substrings included and excluded in the app name).

## Overview

Need to make some common changes across multiple devtron applications?
**Bulk Edit** allows you to do that.<br>
Eg. You can change the value for `MaxReplicas` in Deployment Templates of multiple Devtron applications.

## Support
Bulk edit is currently supported for:
 - Deployment Template

_Ability to edit more Devtron components will be added in the future._

## Steps:

1. Click on the `Bulk Edit` option in the main navigation. This is where you can write and execute scripts to perform bulk updates in Devtron objects.
 
![](../.gitbook/assets/bulk-update-empty.png)
<br>

2. To help you get started, a script template is provided under the `See Samples` section.

![](../.gitbook/assets/bulk-update-readme.png)
<br>

3. Copy and Paste the `Sample Script` in the code editor and make desired changes. Refer `Payload Configuration` in the Readme to understand the parameters.


![](../.gitbook/assets/bulk-update-script.png)

### Example
Example below will select all applications having `abc or xyz` in their name and out of those will exclude applications having `abcd or xyza` in their name. Since global flag is false and envId 23 is provided, it will make changes in envId 23 and not in global deployment template for the selected applications.
If you want to update global deployment template then please set `global: true`. You can set both `global: true` and envIds in the script. If you have provided envId but deployment template is not overridden for that particular environment then it will not apply the changes.



### Sample Script

This is the piece of code which works as the input and has to be pasted in the code editor for achieving bulk updation
task.

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
      patchJson: '[{ "op": "add", "path": "/MaxSurge", "value": 1 },{"op": "replace","path":"/GracePeriod","value": "30"}]'
```


### Payload Configuration
The following tables list the configurable parameters of the Payload component in the Script and their description along with example.
| Parameter                      | Description                        | Example                                                    |
| -------------------------- | ---------------------------------- | ---------------------------------------------------------- |
|`includes.names `        | Will filter apps having exact string or similar substrings                 | `["app%","%abc", "xyz"]` (will include all apps having `"app%"` **OR** `"%abc"` as one of their substring, example - app1, app-test, test-abc etc. **OR** application with name xyz)    |
| `excludes.names`          | Will filter apps not having exact string or similar substrings.              | `["%z","%y", "abc"]`       (will filter out all apps having `"%z"` **OR** `"%y"` as one of their substring, example - appz, test-app-y etc. **OR** application with name abc)                                        |
| `envIds`       | List of envIds to be updated for the selected applications           | `[1,2,3]`                                                   |
| `global`       | Flag to update global deployment template of applications            | `true`,`false`                                                        |
| `patchJson`      | String having the update operation(you can apply more than one changes at a time). It supports [JSON patch ](http://jsonpatch.com/) specifications for update. | `''[ { "op": "add", "path": "/MaxSurge", "value": 1 }, { "op": "replace", "path": "/GracePeriod", "value": "30" }]''` |



<br>


4. Once you have modified the script, you can click on the `Show Impacted Objects` button to see the names of all applications that will be modified when the script is `Run`.


![](../.gitbook/assets/bulk-update-impactobj.png)

<br>

5. Click on the `Run` button to execute the script. Status/Output of the script execution will be shown in the `Output` section of the bottom drawer.


![](../.gitbook/assets/bulk-update-run.png)
<br>



