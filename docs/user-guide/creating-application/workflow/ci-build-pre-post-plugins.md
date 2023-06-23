# Pre-Build and Post-Build Stages

The CI pipeline includes Pre and Post-build steps to validate and introduce checkpoints in the build process.
The pre/post plugins allow you to execute some standard tasks, such as Code analysis, Load testing, Security scanning etc. You can build custom pre-build/post-build tasks or select one of the standard preset plugins provided by Devtron.

Preset plugin is an API resource which you can add within the CI build environment. By integrating the preset plugin in your application, it helps your development cycle to keep track of finding bugs, code duplication, code complexity, load testing, security scanning etc. You can analyze your code easily.

> Devtron CI pipeline includes the following build stages:
>
> * Pre-Build Stage: The tasks in this stage run before the image is built.
> * Build Stage: In this stage, the build is triggered from the source code (container image) that you provide.
> * Post-Build Stage: The tasks in this stage are triggered once the build is complete.

## Before you begin

Make sure you have [CI build pipeline](./ci-pipeline.md) before you start configuring Pre-Build or Post-Build tasks.

## Configuring Pre/Post-build Tasks

Each Pre/Post-build stage is executed as a series of events called tasks and includes custom scripts.
You can create one or more tasks that are dependent on one another for execution. In other words, the output variable of one task can be used as an input for the next task to build a CI runner. The tasks will run following the execution order. 

The tasks can be re-arranged by drag-and-drop; however, the order of passing the variables must be followed.

You can create a task either by selecting one of the available preset plugins or by creating a custom script.

| Stage | Task |
| :--- | :--- |
| Pre-Build/Post-Build | <ol><li>Create a task using one of the [Preset Plugins](#preset-plugins) integrated in Devtron:<ul><li>[K6 Load testing](#k6-load-testing)</li><li>[Sonarqube](#sonarqube)</li><li>[Dependency track for Python](#dependency-track-for-python)</li><li>[Dependency track for NodeJs](#dependency-track-for-nodejs)</li><li>[Dependency track for Maven and Gradle](#dependency-track-for-maven--gradle)</li><li>[Semgrep](#semgrep)</li><li>[Codacy](#codacy)</li></ul></li><li>Create a task from [Execute Custom script](#execute-custom-script) which you can customize your script with:<ul><li>[Custom script - Shell](#custom-script-shell)</li><li>Or, [Custom script - Container image](#custom-script-container-image)</li></ul></li></ol> | 


## Creating Pre/Post-build Tasks

Lets take `Codacy` as an example and configure it in the Pre-Build stage in the CI pipeline for finding bugs, detecting dependency vulnerabilities, and enforcing code standards.

* Go to the **Applications** and select your application from the **Devtron Apps** tabs.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/plugins-latest/applications-app.jpg)


* Go to the **App Configuration** tab, click **Workflow Editor**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/plugins-latest/app-configuration.jpg)


* Select the build pipeline for configuring the pre/post-build tasks.
* On the **Edit build pipeline**, in the `Pre-Build Stage`, click **+ Add task**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/plugins-latest/add-task-pre-build-stage.jpg)


* Select **Codacy** from **PRESET PLUGINS**.
* Enter a relevant name or codacy in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field. <br>`Note`: The description is available by default.
* In the **Input Variables**, provide the information in the following fields:

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/plugins-latest/codacy-details-pre-build-stage.jpg)

 | Variable | Format | Description |
| ---- | ---- | ---- |
| CodacyEndpoint | String | API endpoint for Codacy. |
| GitProvider | String | Git provider for the scanning. |
| CodacyApiToken | String | API token for Codacy. If it is provided, it will be used, otherwise it will be picked from Global secret (CODACY_API_TOKEN). |
| Organisation | String | Your Organisation for Codacy.|
| RepoName | String | Your Repository name. |
| Branch | String | Your branch name. |

* In `Trigger/Skip Condition`, set the trigger conditions to execute a task or `Set skip conditions`. As an exmple: CodacyEndpoint equal to https://app.codacy.com.<br>`Note`: You can set more than one condition.

* In `Pass/Failure Condition` set the conditions to execute pass or fail of your build. As an example: Pass if number of issues equal to zero. <br>`Note`: You can set more than one condition.

* Click **Update Pipeline**.

* Go to the **Build & Deploy**, click the build pipiline and start your build.

* Click `Details` on the build pipeline and you can view the details on the `Logs`.



### Preset plugins

#### K6 Load Testing

K6 is an open-source tool and cloud service that makes load testing easy for developers and QA engineers.

**Prerequisite**: Make sure you have set up an account in `k6.io` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **K6 Load Testing** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| RelativePathToScript | String | Checkout path + script path along with script name. |
| PrometheusUsername | String | Username of Prometheus account. |
| PrometheusApiKey | String | API key of Prometheus account. |
| PrometheusRemoteWriteEndpoint | String | Remote write endpoint of Prometheus account. |
| OutputType | String | `Log` or `Prometheus` |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.

#### Sonarqube

Configuring `Sonarqube` in pre-build or post build task enhances your workflow with Continuous Code Quality & Code Security.

**Prerequisite**: Make sure you have set up an account in `Sonarqube` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Sonarqube** from **PRESET PLUGINS**.

![](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/preset-plugin-sonarqube-2.png)

* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| SonarqubeProjectKey | String | Project key of sonarqube account. |
| SonarqubeApiKey | String | API key of Sonarqube account. |
| SonarqubeEndpoint | String | API endpoint of Sonarqube account. |
| CheckoutPath | String | Checkout path of Git material. |
| UsePropertiesFileFromProject | Boolean | Enter either `true` or `false` accordingly whether you want the poll to generate report or not. |
| CheckForSonarAnalysisReport | Boolean | Enter either `true` or `false` accordingly whether you want the poll to generate report or not. |
| AbortPipelineOnPolicyCheckFailed | Boolean | Enter either `true` or `false` accordingly whether you want to check if the policy fails or not. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.


#### Dependency Track for Python

Configuring `Dependency Track for Python` in pre-build or post build task creates a bill of materials from Python projects and environments and uploads it to D-track for [Component Analysis](https://owasp.org/www-community/Component_Analysis) to identify and reduce risk in the software supply chain.

**Prerequisite**: Make sure you have set up an account in `dependency track` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Dependency track for Python** from **PRESET PLUGINS**.



* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| ProjectManifestType | String | Type of your Python project manifest which is used to build cycloneDx [Software Bill of Materials (SBOM)](https://owasp.org/www-community/Component_Analysis#software-bill-of-materials-sbom). E.g., PIP, Poetry etc. |
| RelativePathToPoetryLock | String | Path to your poetry.lock file inside your project. |
| RelativePathToPipfile | String | Path to your Pipfile.lock file inside your project. |
| RelativePathToRequirementTxt | String | Path to your requirements.txt file inside your project. |
| DTrackEndpoint | String | API endpoint of your dependency track account. |
| DTrackProjectName | String | Name of your dependency track project. |
| DTrackProjectVersion | String | Version of dependency track project. |
| DTrackApiKey | String | API key of your dependency track account. |
| CheckoutPath | String | Checkout path of Git material. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.


#### Dependency Track for NodeJs

Configuring `Dependency Track for NodeJs` in pre-build or post build task creates a bill of materials from NodeJs projects and environments and uploads it to D-track for [Component Analysis](https://owasp.org/www-community/Component_Analysis) to identify and reduce risk in the software supply chain.


**Prerequisite**: Make sure you have set up an account in `dependency track` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Dependency track for NodeJs** from **PRESET PLUGINS**.



* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| DTrackEndpoint | String | API endpoint of your dependency track account. |
| DTrackProjectName | String | Name of your dependency track project. |
| DTrackProjectVersion | String | Version of dependency track project. |
| DTrackApiKey | String | API key of your dependency track account. |
| CheckoutPath | String | Checkout path of Git material. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.



#### Dependency Track for Maven & Gradle

Configuring `Dependency Track for NodeJs` in pre-build or post build task creates a bill of materials from Maven & Gradle projects and environments and uploads it to D-track for [Component Analysis](https://owasp.org/www-community/Component_Analysis) to identify and reduce risk in the software supply chain.


**Prerequisite**: Make sure you have set up an account in `dependency track` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Dependency track for Maven & Gradle** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| BuildToolType | String | Type of build tool your project is using. E.g., Maven, or Gradle. |
| DTrackEndpoint | String | API endpoint of your dependency track account. |
| DTrackProjectName | String | Name of your dependency track project. |
| DTrackProjectVersion | String | Version of dependency track project. |
| DTrackApiKey | String | API key of your dependency track account. |
| CheckoutPath | String | Checkout path of Git material. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.


#### Semgrep

Semgrep is a fast, open source, static analysis engine for finding bugs, detecting dependency vulnerabilities, and enforcing code standards.

**Prerequisite**: Make sure you have set up an account in `Semgrep` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Semgrep** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| SemgrepAppToken | String | App token of Semgrep. If it is provided, this token will be used, otherwise it will be picked from Global Secret. |
| PrefixAppNameInSemgrepBranchName | Bool | Enter either `true` or `false` accordingly whether you want app name to be reflected with a branch name. If it is `true`, it will add app name with branch name. E.g., {SemgrepAppName}-{branchName} |
| UseCommitAsSemgrepBranchName | Bool | Enter either `true` or `false` accordingly whether you want app name to be reflected with commit hash. If it is `true`, it will add app name with commit hash. E.g., {SemgrepAppName}-{CommitHash}. |
| SemgrepAppName | String | App name for Semgrep. If it is provided, and `PrefixAppNameInSemgrepBranchName` is true, then this will be prefixed with branch name/commit hash.|
| ExtraCommandArguments | String | Extra command arguments for Semgrep CI command. E.g., Input: --json --dry-run. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.



#### Jira Issue Validator During Git Push

Jira is a tool developed for teams to track issues, manage projects, and automate workflows. This plugin extends the filtering capabilities of the Devtron CI and lets the users perform validation based on Jira Ticket ID status.

**Prerequisite**: Make sure you have the Jira issue title and ticket-id.<br>
Note: During build pipeline must choose source type to `Pull Request` and configure the webhook that the title of Pull request ` ^(?P<Jira_Id>([a-zA-Z0-9-].*))/ ` should be in this regular expression format .</br>
Example-: If your jira id is DEMO-1 then your pull request title  should be like DEMO-1/any title for pr.

1. On the **Edit Build Pipeline** page, select the **Pre-Build Stage**.
2. Click **+ Add task**.
3. Select **Jira Issue Validator** from **PRESET PLUGINS**.

* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide values for the input variables. It is a mandatory field.

 | Variable | Format | Description |
| ---- | ---- | ---- |
| JiraUsername | String | Enter the  username of Jira account.  |
| JiraPassword | String | Enter the api-token for Jira account. Don't write the password of your Jira account  |
| JiraBaseUrl  | String | Enter the base url of your Jira account. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.

#### Jira Issue Updater After build

This plugin extends the capabilities of Devtron CI and can update issues in Jira by adding pipeline status and metadata as comment on the tickets.

**Prerequisite**: Make sure you have the Jira issue title or ticket-id.

1. On the **Edit Build Pipeline** page, select the **Post-Build Stage**.
2. Click **+ Add task**.
3. Select **Jira Issue Updater** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide values for the input variables. It is a mandatory field.

 | Variable | Format | Description |
| ---- | ---- | ---- |
| JiraUsername | String | Enter the username of Jira account.  |
| JiraPassword | String | Enter the api-token for Jira account. Don't write the password of your Jira account.  |
| JiraBaseUrl  | String | Enter the base url of your Jira account. |
| UpdateWithDockerImageId | Boolean | Enter either `true` or `false`. If it is `true`, Jira issue will be updated with docker image id in comment. By default it is true. |
| UpdateWithBuildStatus | Boolean |  Enter either `true` or `false`. If it is `true`, Jira issue will be updated with build status in comment. By default it is true. |
* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* Click **Update Pipeline**.


#### Codacy

Codacy is an automated code analysis/quality tool that helps developers to ship better software in a faster manner.

**Prerequisite**: Make sure you have set up an account in `Codacy` or get the API keys from an admin.

1. On the **Edit build pipeline** page, select the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Codacy** from **PRESET PLUGINS**.


* Enter a relevant name in the `Task name` field. It is a mandatory field.
* Enter a descriptive message for the task in the `Description` field. It is an optional field.
* Provide a value for the input variable.<br> Note: The value may be any of the values from the previous build stages, a global variable, or a custom value.</br>

 | Variable | Format | Description |
| ---- | ---- | ---- |
| CodacyEndpoint | String | API endpoint for Codacy. |
| GitProvider | String | Git provider for the scanning. |
| CodacyApiToken | String | API token for Codacy. If it is provided, it will be used, otherwise it will be picked from Global secret (CODACY_API_TOKEN). |
| Organisation | String | Your Organisation for Codacy.|
| RepoName | String | Your Repository name. |
| Branch | String | Your branch name. |

* `Trigger/Skip Condition` refers to a conditional statement to execute or skip the task. You can select either:<ul><li>`Set trigger conditions` or</li><li>`Set skip conditions`</li></ul> 

* `Pass/Failure Condition` refers to conditions to execute pass or fail of your build. You can select either:<ul><li>`Set pass conditions` or</li><li>`Set failure conditions`</li></ul> 

* Click **Update Pipeline**.


### Execute custom script

1. On the **Edit build pipeline** screen, select the **Pre-build stage**.
2. Select **+ Add task**.
3. Select **Execute custom script**.

The task type of the custom script may be a [Shell](#custom-script---shell) or a [Container image](#custom-script---container-image).

#### Custom script - Shell

* Select the **Task type** as **Shell**.

Consider an example that creates a Shell task to stop the build if the database name is not "mysql". The script takes 2 input variables, one is a global variable (`DOCKER_IAMGE`), and the other is a custom variable (`DB_NAME`) with a value "mysql".
The task triggers only if the database name matches "mysql".
If the trigger condition fails, this Pre-build task will be skipped and the build process will start.
The variable `DB_NAME` is declared as an output variable that will be available as an input variable for the next task.
The task fails if `DB_NAME` is not equal to "mysql".

![Custom script - Shell](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/Custom-script-Shell.png)

| Field name | Required/Optional | Field description |
| --- | --- | --- |
| Task name | Required | A relevant name for the task |
| Description | Optional | A descriptive message for the task |
| Task type | Optional | Shell: Custom shell script goes here |
| Input variables | Optional | <ul><li>**Variable name**: Alphanumeric chars and (_) only</li><li>**Source or input value**: The variable's value can be global, output from the previous task, or a custom value. <br>Accepted data types include: STRING \| BOOL \| NUMBER \| DATE</li><li>**Description**: Relevant message to describe the variable.</li></ul> | The input variables will be available as env variables |
| Trigger/Skip condition | Optional | A conditional statement to execute or skip the task |
| Script | Required | Custom script for the Pre/Post-build tasks |
| Output directory path | | Optional | Directory path for the script output files such as logs, errors, etc. |
| Output variables | Optional | Environment variables that are passed as input variables for the next task. <ul><li>Pass/Failure Condition (Optional): Conditional statements to determine the success/failure of the task. A failed condition stops the execution of the next task and/or build process</li></ul> |
 
* Select **Update Pipeline**.

Here is a screenshot with the failure message from the task:

![Pre-Build task failure](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/custom-script-shell-task-failed.png)

#### Custom script - Container image

* Select the **Task type** as **Container image**.

This example creates a Pre-build task from a container image. The output variable from the previous task is available as an input variable.

![Custom script - Container image](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/Custom-script-Container-image.png)

| Field name | Required/Optional | Field description |
| --- | --- | --- |
| Task name | Required | A relevant name for the task |
| Description | Optional | A descriptive message for the task |
| Task type | Optional | Container image |
| Input variables | Optional | <ul><li>**Variable name**: Alphanumeric chars and (_) only</li><li>**Source or input value**: The variable's value can be global, output from the previous task, or a custom value <br>Accepted data types include: STRING \| BOOL \| NUMBER \| DATE</li><li>**Description**: Relevant message to describe the variable</li></ul> | The input variables will be available as env variables |
| Trigger/Skip condition | Optional | A conditional statement to execute or skip the task |
| Container image | Required | Select an image from the drop-down list or enter a custom value in the format `<image>:<tag>` |
| Mount custom code | Optional | Enable to mount the custom code in the container. Enter the script in the box below. <ul><li>Mount above code at (required): Path where the code should be mounted</li></ul> |
| Command | Optional | The command to be executed inside the container |
| Args | Optional | The arguments to be passed to the command mentioned in the previous field |
| Port mapping | Optional | The port number on which the container listens. The port number exposes the container to outside services |
| Mount code to container | Optional | Mounts the source code inside the container. Default is "No". If set to "Yes", enter the path  |
| Mount directory from host | Optional | Mount any directory from the host into the container. This can be used to mount code or even output directories |
| Output directory path | Optional | Directory path for the script output files such as logs, errors, etc. |
  
* Select **Update Pipeline**.

## What's next

Trigger the [CI pipeline](../../deploying-application/triggering-ci.md)
