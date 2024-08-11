# Image-Scan

## Introduction
The **Image Scan** plugin by Devtron enables you to scan and detect vulnerabilities in container images using Trivy. The Image Scan plugin is recommended to be integrated into the Job Pipeline, especially when you are using external image sources like Jenkins. Based on Image Scan results, you can enforce security policies to either proceed with or abort the deployment process, giving you more control over your deployment process.

### Prerequisites
Before integrating the Image Scan plugin, ensure that you have installed the `Vulnerability Scanning (Trivy/Clair)` integration from Devtron Stack Manager.

---

## Steps
1. Go to **Applications** → **Devtron Apps**.
2. Click your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **CREATE JOB PIPELINE**.
5. Enter the required fields in the **Basic configuration** window.
6. Click **Task to be executed**.
7. Under 'TASKS', click the **+ Add task** button.
8. Click the **Image Scan** plugin.
9. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task

e.g., `Scanning container image`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The Image Scan plugin is integrated for detecting vulnerabilities in container image.`

### Input Variables

No input variables are required for Image Scan plugin.

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Image Scan will not be generating an output variable.

Click **Update Pipeline**.



