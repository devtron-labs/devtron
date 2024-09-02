# Code Scan

## Introduction
The Code Scan plugin of Devtron allows you to perform the code scanning using Trivy. By integrating the **Code Scan** plugin into your workflow you can detect common Vulnerabilities, Misconfigurations, License Risks, and Exposed Secrets in your code.

### Prerequisites
Before integrating the **Code Scan** plugin, install the [Vulnerability Scanning (Trivy/Clair)](https://docs.devtron.ai/usage/integrations/clair) integration from Devtron Stack Manager.

---

## Steps
1. Go to **Applications** → **Devtron Apps**.
2. Click your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **Build and Deploy from Source Code**.
5. Fill the required fields in the **Create build pipeline** window and navigate to the **Pre-build stage**.

{% hint style="warning" %}
If you have already configured workflow, edit the build pipeline, and navigate to **Pre-build stage**.
{% endhint %}

6. Under 'TASKS', click the **+ Add task** button.
7. Select the **Code Scan** plugin.
8. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task 

e.g., `Code Scanning`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

e.g., `The Code Scan plugin is integrated for scanning the in-code vulnerabilities.`

### Input Variables

No input variables are required for the Code Scan plugin.

### Output Variables
Code Scan will not be generating an output variable.

Click **Update Pipeline**.


