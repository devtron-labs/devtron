# Jenkins

## Introduction
Jenkins is an open-source Continuous Integration (CI) server. You can manage multiple stages of software delivery using Jenkins including Automated testing, Static Code Analysis, Building, Packaging, and Deploying.
With Devtron's Jenkins plugin, you can: 
- Trigger pre-configured Jenkins jobs from Devtron and stream the logs to the Devtron dashboard.
- Execute Jenkins build pipelines through Devtron and deploy to target environments using the Devtron CD pipeline.

### Prerequisites
Before integrating the Jenkins plugin, ensure that you have properly configured your Jenkins job and also have the necessary parameters set for triggering from Devtron.

---

## Steps
1. Go to Applications → **Devtron Apps**.
2. Click on your application.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click **New Workflow** and navigate to the **CREATE JOB PIPELINE**.
5. Enter the required fields in the **Basic configuration** window.
6. Under 'TASKS', click the **+ Add task** button.
7. Click the **Jenkins plugin**.
8. Enter the following [user inputs](#user-inputs) with appropriate values.
---

## User Inputs

### Task Name
Enter the name of your task.

`e.g. Jenkins_Job`

### Description
Add a brief explanation of the task and the reason for choosing the plugin. Include information for someone else to understand the purpose of the task.

`e.g. Trigger the build Job of Jenkins`

### Input Variables

| Variable                 | Format       | Description | Sample Value |
| ------------------------ | ------------ | ----------- | ------------ |
|   URL                    | STRING       | The base URL of the Jenkins server.            | https://jenkins.example.com             |
|   USERNAME               | STRING       | Username for Jenkins server.            | admin |
|   PASSWORD               | STRING       | Password of the Jenkins user for authentication             | securePass123!             |
|   JOB_NAME               | STRING       | The name of the Jenkins job to be triggered.           | CI-build-job             |
|   JOB_TRIGGER_PARAMS     | STRING       | Parameters to be passed for triggering a job.            | branch=main&environment=production            |
|   JENKINS_PLUGIN_TIMEOUT | INTEGER       | The maximum time (in minutes) to wait for a Jenkins plugin operation to complete before timing out.            |  60            |

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Jenkins will not be generating an output variable.

Click **Update Pipeline**.


