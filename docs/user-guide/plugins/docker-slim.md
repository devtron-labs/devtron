# DockerSlim

## Introduction
The **DockerSlim** plugin by Devtron helps you to optimize your container deployments by reducing Docker image size. Now with these lighter Docker images, you can perform faster deployments and enhance overall system efficiency. 

{% hint style="warning" %}
Support for Docker buildx images will be added soon.
{% endhint %}

### Prerequisites
No prerequisites are required for integrating the **DockerSlim** plugin.

---

## Steps
1. Go to **Applications** → **Devtron Apps**.
2. Click your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **Build and Deploy from Source Code**.
5. Fill the required fields in the **Create build pipeline** window and navigate to the **Post-build stage**.

{% hint style="warning" %}
If you have already configured workflow, edit the build pipeline, and navigate to **Post-build stage**.
{% endhint %}

6. Under 'TASKS', click the **+ Add task** button.
7. Click the **DockerSlim** plugin.
8. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task

e.g., `Reduce Docker image size`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The DockerSlim plugin is integrated for reducing the size of Docker image.`

### Input Variables

{% hint style="warning" %}
At `IncludePathFile` input variable list down the file path of essential files from your Dockerfile. Files for which the path is not listed  at `IncludePathFile` will may be excluded from the Docker image to reduce size.
{% endhint %}

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   HTTPProbe              | BOOL         | Indicates whether the port is exposed in Dockerfile or not | false                         |
|   IncludePathFile        | STRING       | File path of required files            | /etc/nginx/include.conf       |

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
DockerSlim will not be generating an output variable.

Click **Update Pipeline**.



