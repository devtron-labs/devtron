# Pre-Build and Post-Build stages

The CI pipeline includes Pre and Post-build steps to validate and introduce checkpoints in the build process.

The pre/post plugins allow you to execute some standard tasks, such as Code analysis, Load testing, Security scanning, and so on.
You can build custom pre/post tasks or use one from the standard preset plugins provided by Devtron.

## Before you begin

Create a [CI build pipeline](./ci-pipeline2.md) if you haven't done that already!

## Configuring Pre/Post-build stages

Each Pre/Post-build stage is executed as a series of events called tasks and includes custom scripts.
You could create one or more tasks that are dependent on one another for execution. In other words, the output variable of one task can be used as an input for the next task to build a CI runner.

The tasks will run following the execution order.

> The tasks may be re-arranged by using drag-and-drop; however, the order of passing the variables must be followed.

| Stage | Task |
| :--- | :--- |
| Pre-Build/Post-Build | <ol><li>Create a task from - [Preset Plugin](#preset-plugins)<ul><li>Sonarqube</li><li>K6 Load testing</li></ul></li><li>Create a task from - [Execute Custom script](#execute-custom-script)<ul><li>[Custom script - Shell](#custom-script-shell)</li><li>[Custom script - Container image](#custom-script-container-image)</li></ul></li></ol> | 

## Creating a task

1. Go to **Applications** and select your application from the **Devtron Apps** tabs.
2. From the **App Configuration** tab select **Workflow Editor**.
3. Select the build pipeline for editing the stages.

> Devtron CI pipeline includes the following build stages:
>
> * Pre-build stage: The tasks in this stage run before the image is built.
> * Build stage: In this stage, the build is triggered from the source code that you provide.
> * Post-build stage: The tasks in this stage are triggered once the build is complete.

You can create a task either by selecting one of the available preset plugins or by creating a custom script.

[Preset plugins](#preset-plugins) | [Execute custom script](#execute-custom-script)

### Preset plugins

**Prerequisite**: Set up Sonarqube, or get the API keys from an admin.

The example shows a Post-build stage with a task created using a preset plugin - Sonarqube.

1. On the **Edit build pipeline** screen, select the **Post-build stage** (or Pre-build).
2. Select **+ Add task**.
3. Select **Sonarqube** from **PRESET PLUGINS**.

![Preset plugin - Sonarqube](https://devtron-public-asset.s3.us-east-2.amazonaws.com/plugins/preset-plugin-sonarqube.png)

| Field name | Required/Optional | Field description |
| --- | --- | --- |
| Task name | Required  | A relevant name for the task |
| Description | Optional | A descriptive message for the task |
| Input variables | Optional | VALUE: A value for the input variable. The value may be any of the values from the previous build stages, a global variable, or a custom value |
| Trigger/Skip Condition | Optional | A conditional statement to execute or skip the task |

Select **Update Pipeline**.

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
