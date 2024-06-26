# Create Your Plugin

## Introduction

You can create CI/CD plugins using APIs. It can be any of the following: CI plugin or CD plugin.

Your plugin can be a single-step or multi-step plugin, where steps can be considered as tasks. The task can either be simple shell commands or it can be complex operations that require a specific container environment.

---

## API Call

{% hint style="warning" %}
### Prerequisite
You will need a [token](../../user-guide/global-configurations/authorization/api-tokens.md) to make API calls
{% endhint %}

```
POST {{DEVTRON_BASEURL}}/orchestrator/plugin/global
```

---

## Example Plugin

In the following example, we are creating a single-step plugin named **Secret Management Validator**. Moreover, we want to execute a simple shell script; therefore, we are keeping the task type as `SHELL`

### Sample Request Body

{% code title="Plugin Request Body" overflow="wrap" lineNumbers="true" %}

```json
{
    "name": "Secret Management Validator",
    "description": "The Secret Management Validator plugin integrates with your CI/CD pipeline to automatically detect and prevent the inclusion of secrets or sensitive information in your codebase, ensuring compliance and security.",
    "type": "SHARED",
    "icon": "https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/vectors/secret-management-validator.png",
    "tags": ["security", "compliance", "secrets"],
    "action": 0,
    "pluginStage": "CI_CD",
    "pluginSteps": [
        {
            "name": "Step 1",
            "description": "Step 1 - Secret Management Validator",
            "index": 1,
            "stepType": "INLINE",
            "refPluginId": 0,
            "outputDirectoryPath": null,
            "dependentOnStep": "",
            "pluginStepVariable": [
                {
                    "name": "PathToScan",
                    "format": "STRING",
                    "description": "The relative path to the directory or file that needs to be scanned for secrets.",
                    "isExposed": true,
                    "allowEmptyValue": true,
                    "defaultValue": "",
                    "variableType": "INPUT",
                    "valueType": "NEW",
                    "variableStepIndex": 1,
                    "variableStepIndexInPlugin": 0
                },
                {
                    "name": "GitGuardianApiKey",
                    "format": "STRING",
                    "description": "The API key for GitGuardian to authenticate and use the secret detection service.",
                    "isExposed": true,
                    "allowEmptyValue": false,
                    "defaultValue": "",
                    "variableType": "INPUT",
                    "valueType": "NEW",
                    "variableStepIndex": 1,
                    "variableStepIndexInPlugin": 0
                },
                {
                    "name": "ScanScope",
                    "format": "STRING",
                    "description": "Defines the scope of the scan. It can be set to scan all files, specific file types, or based on patterns.",
                    "isExposed": true,
                    "allowEmptyValue": true,
                    "defaultValue": "all",
                    "variableType": "INPUT",
                    "valueType": "NEW",
                    "variableStepIndex": 1,
                    "variableStepIndexInPlugin": 0
                },
                {
                    "name": "OutputFormat",
                    "format": "STRING",
                    "description": "The desired format for the output report, such as JSON, HTML, or plaintext.",
                    "isExposed": true,
                    "allowEmptyValue": true,
                    "defaultValue": "JSON",
                    "variableType": "INPUT",
                    "valueType": "NEW",
                    "variableStepIndex": 1,
                    "variableStepIndexInPlugin": 0
                }
            ],
            "pluginPipelineScript": {
                "script": "\n# Run GitGuardian secret detection\nif [ -n \"$GITGUARDIAN_API_KEY\" ]; then\n echo \"Running GitGuardian Secret Detection...\"\n ggshield scan path $SCAN_PATH --api-key $GITGUARDIAN_API_KEY\nelse\n echo \"GitGuardian API key is missing. Skipping secret detection.\"\nfi\n\n# Output the results\nif [ -f ggshield-output.json ]; then\n cat ggshield-output.json\nelse\n echo \"No GitGuardian output found.\"\nfi",
                "storeScriptAt": "",
                "type": "SHELL"
            }
        }
    ]
}

```
{% endcode %}

Required fields to edit in the above sample payload are:

| Key Path     | Description                                                   |
| ------------ | ------------------------------------------------------------- |
| name         | Plugin name                                                   |
| description  | Plugin description                                            |
| tags         | Array of tags                                                 |
| icon         | Plugin icon url                                               |
| Plugin steps | Array of tasks to execute (Details of fields discussed below) |

Fields of a plugin steps are:

| Key Path                    | Description                                  |
| --------------------------- | -------------------------------------------- |
| name                        | Step name                                    |
| description                 | Description of step                          |
| index                       | Sequence at which the step needs to executed |
| outputDirectoryPath         | Artifact output path                         |
| pluginStepVariable          | Array of required input / output variables   |
| pluginPipelineScript.script | Stringified bash script                      |


### Result

Your new plugin will appear under **Shared Plugins** depending on which stage you have created it for: pre/post build (`pluginStage = CI`), pre/post deployment (`pluginStage = CD`), or both (`pluginStage = CI_CD`)

![New Shared Plugin](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/plugins/create-plugin/shared-plugin.jpg)

The variables defined in the `pluginStepVariable` array would appear as shown below.

![Plugin Fields](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/plugins/create-plugin/plugin-fields.jpg)

---

## Other API calls

To fetch details of a specific plugin by its ID

```
GET 
/orchestrator/plugin/global/detail/{pluginId}
```

To fetch details of all plugins

```
GET
/orchestrator/plugin/global/detail/all
```

To fetch list of all global variables

```
GET
/orchestrator/plugin/global/list/global-variable
```

---

## Field Definitions

Refer the [spec file](https://github.com/devtron-labs/devtron/blob/main/specs/global-plugin.yaml) for detailed definition of each field present in the request/response body of the API.
