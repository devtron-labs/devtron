# Devtron-Job-Trigger

## Introduction
The **Devtron Job Trigger** plugin enables you to trigger Devtron Jobs from your current application workflow. For example, by integrating this plugin at the pre-deployment stage of your application workflow, you can trigger jobs designed to run migration scripts in your database. This ensures that necessary migrations are executed before your application is deployed.

### Prerequisites
Before integrating the Devtron Job Trigger plugin, you need to properly configure the target Devtron Job to ensure smooth execution.

---

## Steps
1. Go to **Applications** → **Devtron Apps**.
2. Click your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **Build and Deploy from Source Code**.
5. Fill the required fields in the **Create build pipeline** window and navigate to the **Create deployment pipeline**.
6. Fill the required fields in the **Deployment Stage** window and navigate to the **Pre-Deployment stage**.

{% hint style="warning" %}
If you have already configured workflow, edit the deployment pipeline, and navigate to **Pre-Deployment stage**.
{% endhint %}

6. Under 'TASKS', click the **+ Add task** button.
7. Select the **Devtron Job Trigger** plugin.
8. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task

e.g., `Triggers Devtron Job `

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The Devtron Job Trigger plugin is integrated for triggering the Devtron Job.`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   DevtronApiToken        | STRING       | Enter Devtron API token with required permissions. | abc123def456token789            |
|   DevtronEndpoint        | STRING       | Enter the URL of Devtron dashboard.     | https://devtron.example.com            |
|   DevtronJob             | STRING       | Enter the name or ID of Devtron Job to be triggered  | plugin-test-job |
|   DevtronEnv             | STRING       | Enter the name or ID of the Environment where the job is to be triggered. If JobPipeline is given, ignore this field and do not assign any value |      prod     |
|   JobPipeline            | STRING       | Enter the name or ID of the Job pipeline to be triggered. If DevtronEnv is given, ignore this field and do not assign any value  | hello-world  |
|   GitCommitHash          | STRING       | Enter the commit hash from which the job is to be triggered. If not given then, will pick the latest  |    cf19e4fd348589kjhsdjn092nfse01d2234235sdsg        |
|   StatusTimeoutSeconds   | NUMBER       | Enter the maximum time to wait for the job status |   120   |

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Devtron Job Trigger will not be generating an output variable.

Click **Update Pipeline**.



