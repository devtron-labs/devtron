# Devtron-CD-Trigger

## Introduction
The **Devtron CD Trigger** plugin allows you to trigger the PRE-CD, CD, or POST-CD stages of target Devtron App from within your current application workflow. This plugin offers flexibility in managing application dependencies and deployment sequences. For example, by incorporating this plugin at the pre-deployment stage of your application workflow, you can deploy another application that contains dependencies required by your current application, ensuring a coordinated deployment process.

### Prerequisites
Before integrating the Devtron CD Trigger plugin, you need to properly configure the target Devtron App to ensure smooth execution.

---

## Steps
1. Go to **Applications** → **Devtron Apps**.
2. Click your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **Build and Deploy from Source Code**.
5. Fill the required fields in the **Create build pipeline** window and navigate to the **Create deployment pipeline**.
6. Fill the required fields in the **Deployment Stage** window and navigate to the **Post-Deployment stage**.

{% hint style="warning" %}
If you have already configured workflow, edit the deployment pipeline, and navigate to **Post-Deployment stage**.
{% endhint %}

6. Under 'TASKS', click the **+ Add task** button.
7. Select the **Devtron CD Trigger** plugin.
8. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task

e.g., `Triggers CD Pipeline`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The Devtron CD Trigger plugin is integrated for triggering the CD stage of another application.`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   DevtronApiToken        | STRING       | Enter target Devtron API token. |  abc123DEFxyz456token789            |
|   DevtronEndpoint        | STRING       | Enter the target URL of Devtron.     | https://devtron.example.com            |
|   DevtronApp             | STRING       | Enter the target Devtron Application name/ID | plugin-demo |
|   DevtronEnv             | STRING       | Enter the target Environment name/ID. Required if JobPipeline is not given |  preview         |
|   StatusTimeoutSeconds           | STRING       | Enter the maximum time (in seconds) a user can wait for the application to deploy. Enter a positive integer value   | 120  |
|   GitCommitHash          | STRING       | Enter the git hash from which user wants to deploy its application. By default it takes latest Artifact ID to deploy the application |    cf19e4fd348589kjhsdjn092nfse01d2234235sdsg        |
|   TargetTriggerStage   | STRING       | Enter the Trigger Stage PRE/DEPLOY/POST. Default value is `Deploy`. |   PRE   |

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Devtron CD Trigger will not be generating an output variable.

Click **Update Pipeline**.



