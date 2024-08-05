# Copacetic

## Introduction
The Copacetic plugin of Devtron helps you patch your container image vulnerabilities traced by the security scan Devtron performed on your container image. By integrating the **Copacetic** plugin into your workflow and enabling the **Scan for vulnerabilities** at your **Build stage**, you can:
- Trace the vulnerabilities of your container images, and the **Copacetic** plugin will automatically patch the container image vulnerabilities for you.

### Prerequisites
Before integrating the **Copacetic** plugin, install the `Vulnerability Scanning (Trivy/Clair)` integration from Devtron Stack Manager. Once the integration is installed, make sure you have enabled **Scan for vulnerabilities** at the **Build stage** or integrated the [Code-Scan](./code-scan.md) plugin in the **Pre-build stage**.

---

## Steps
1. Go to **Applications** → **Devtron Apps**.
2. Click your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **Build and Deploy from Source Code**.
5. Fill the required fields in the **Create build pipeline** window and navigate to the **Post-build stage**.

{% hint style="info" %}
If you have already configured workflow, edit the build pipeline, and navigate to **Pre-build stage**.
{% endhint %}

6. Under 'TASKS', click the **+ Add task** button.
7. Click the **Copacetic** plugin.
8. Enter the following [user inputs](#user-inputs) with appropriate values.

---

## User Inputs

### Task Name
Enter the name of your task.

e.g., `Patch container image vulnerability`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The Copacetic plugin is configured to patch the vulnerabilities in container image`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   CopaTimeout            | STRING       | Provide timeout for copa patch command, default time is 5 minutes | 10m |


### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Copacetic will not be generating an output variable.

Click **Update Pipeline**.


