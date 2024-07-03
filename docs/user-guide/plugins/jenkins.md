# Jenkins

## Introduction
The Jenkins Plugin for Devtron offers seamless integration of Jenkins jobs into Devtron's CI/CD pipeline workflow. This plugin allows users to trigger external Jenkins jobs directly from the Devtron dashboard, streamlining workflow management. By incorporating this plugin, teams can centralize their CI/CD operations.

### Prerequisites
Before integrating the Jenkins plugin, ensure that you have properly configured your Jenkins jobs.

---

## Steps
1. Go to Applications → **Devtron Apps**.
2. Click on your **application**.
3. Go to **App Configuration** → **Workflow Editor**.
4. Click on your **CI pipeline** and navigate to the **Pre-Build** stage or **Post-Build** stage depending on where you want to integrate the Jenkins plugin.
5. Under 'TASKS', click the **Add task** button.
6. Click the **Jenkins plugin**.
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
|   USERNAME               | STRING       | The username for authenticating with the Jenkins server,with appropriate permissions            | admin_user |
|   PASSWORD               | STRING       | The password for the Jenkins user.             | securePass123!             |
|   JOB_NAME               | STRING       | The name of the Jenkins job to be triggered.           | CI-build-job             |
|   JOB_TRIGGER_PARAMS     | STRING       | Parameters passed when triggering a job.            | branch=main&environment=production            |
|   JENKINS_PLUGIN_TIMEOUT | INTEGER       | The maximum time (in seconds) to wait for a Jenkins plugin operation to complete before timing out.            |  300            |

### Trigger/Skip Condition
Here you can set conditions to execute or skip the task. You can select `Set trigger conditions` for the execution of a task or `Set skip conditions` to skip the task.

### Output Variables
Jenkins will not be generating an output variable.

Click **Update Pipeline**.


