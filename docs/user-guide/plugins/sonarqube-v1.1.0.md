# Sonarqube v1.1.0

Configuring `Sonarqube-v1.1.0` in pre-build or post build task enhances your workflow with Continuous Code Quality & Code Security.

**Prerequisite**: Make sure you have set up an account in `Sonarqube` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Sonarqube v1.1.0** from **PRESET PLUGINS**.

    ![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/sonarqube-v1.1.0.jpeg)

* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| SonarqubeProjectPrefixName | String | This is the SonarQube project prefix name. If not provided, the prefix name is automatically generated. |
| SonarqubeBranchName | String | Branch name to be used to send the scanned result on sonarqube project. |
| SonarqubeProjectKey | String | Project key of SonarQube account |
| CheckForSonarAnalysisReport | Bool | Boolean value - true or false. Set true to poll for generated report from sonarqube. |
| AbortPipelineOnPolicyCheckFailed | Bool | Boolean value - true or false. Set true to abort on report check failed. |
| UsePropertiesFileFromProject | Bool | Boolean value - true or false. Set true to use source code sonar-properties file. |
| SonarqubeEndpoint | String | API endpoint of SonarQube account. |
| CheckoutPath | String | Checkout path of Git material. |
| SonarqubeApiKey | String | API key of SonarQube account |
| SonarContainerImage | String | Container Image that will be used for sonar scanning purpose. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 
* `Pass/Fail Condition` refers to a conditional statement to pass or fail the **Pre-Build Stage** (or Post-Build Stage). You can select either:<ul><li>`Set pass conditions` or</li><li>`Set failure conditions`</li></ul> 

* Click **Update Pipeline**.