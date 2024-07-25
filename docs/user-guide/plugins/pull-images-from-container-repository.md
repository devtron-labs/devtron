# Pull images from container repository

## Introduction
The Pull images from container repository plugin helps you poll the specified container repository and fetch the container images to deploy them on your target Kubernetes environments using Devtron's CD pipeline. By integrating this plugin you can:
- Poll the designated container repository to get the specific container image build using external CI like Jenkins and Github actions. Once the image becomes available, you can deploy it to your target Kubernetes environment using Devtron's CD pipeline.

{% hint style="warning" %}
Currently, this plugin only supports ECR registry, soon support for other container registries will be added.
{% endhint %}

### Prerequisites
Before integrating the **Pull images from the container repository** plugin, ensure that you have a specific container image present at your ECR container repository to pull the image and deploy it to the target environment.

---

## Steps
1. Go to Applications → **Devtron Apps**.
2. Click on your **application**.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click on **New Workflow** and navigate to the **CREATE JOB PIPELINE**.
5. Enter the following fields on the **Basic configuration** window.
6. Click on **Task to be executed** under 'TASKS', click the **Add task** button.
6. Click the **Pull images from container repository** plugin.
7. Enter the following fields with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task.

`e.g. Pull container image`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

`e.g. Pull container image build by external CI`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   REPOSITORY             | STRING       | Provide name of repository for polling | dev-repo |


### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Pull images from container repository will not be generating an output variable.

Click **Update Pipeline**.


