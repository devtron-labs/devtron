# CraneCopy

## Introduction
The **CraneCopy** plugin by Devtron facilitates the transfer of multi-architecture container images between registries. When integrated into Devtron's Post-build stage, this plugin allows you to efficiently copy and store your container images to a specified target repository.

### Prerequisites
No prerequisites are required for integrating the **CraneCopy** plugin.

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
7. Click the **CraneCopy** plugin.
8. Enter the following [user inputs](#user-inputs) with appropriate values.

---

## User Inputs

### Task Name
Enter the name of your task

e.g., `Copy and store container images`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The CraneCopy plugin is integrated to copy the container images from one registry to another.`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   RegistryUsername       |    STRING    | Username of target registry for authentication      |    admin                |
|   RegistryPassword       |    STRING    | Password for the target registry for authentication |    Tr5$mH7p             |
|   TargetRegistry         |    STRING    | The target registry to push to image                |    docker.io/dockertest | 


### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
CraneCopy will not be generating an output variable.

Click **Update Pipeline**.



