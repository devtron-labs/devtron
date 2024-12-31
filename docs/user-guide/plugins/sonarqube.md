# SonarQube

Configuring `Sonarqube` in pre-build or post build task enhances your workflow with Continuous Code Quality & Code Security.

**Prerequisite**: Make sure you have set up an account in `Sonarqube` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Sonarqube** from **PRESET PLUGINS**.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/sonarqube.jpg)

* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| SonarqubeProjectKey | String | Project key of SonarQube account |
| SonarqubeApiKey | String | API key of SonarQube account |
| SonarqubeEndpoint | String | API endpoint of SonarQube account |
| CheckoutPath | String | Checkout path of Git material |
| UsePropertiesFileFromProject | Boolean | Enter either `true` or `false` accordingly whether the configuration file should be fetched from the project's source code |
| CheckForSonarAnalysisReport | Boolean | Enter either `true` or `false` accordingly whether you want poll or actively check for the generation of the SonarQube analysis report |
| AbortPipelineOnPolicyCheckFailed | Boolean | Enter either `true` or `false` accordingly whether you want to check if the policy fails or not |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.