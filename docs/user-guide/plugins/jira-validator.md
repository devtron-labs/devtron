# Jira Issue Validator

## Introduction
The Jira Issue Validator plugin extends the filtering capabilities of the Devtron CI and lets users perform validation based on Jira Ticket ID status. This plugin ensures that only builds associated with valid Jira tickets are executed, improving the accuracy of the CI process.

### Prerequisites

- A Jira account with the necessary [API access](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/#Create-an-API-token).
- The API credentials (username, password, and base URL) for your Jira instance. Obtain the API credentials from your Jira admin if required.
- A pull request raised with your Git provider. Title of pull request must contain the Jira ID.
- Jira Issue (e.g., REDOC-12)
- Webhook added to the git repository. [Click here](https://docs.devtron.ai/usage/applications/creating-application/workflow/ci-pipeline#configuring-webhook) to know more.

---

## Steps

1. On the **Edit build pipeline** page, go to the **Pre-Build Stage** (or Post-Build Stage).
2. Click **+ Add task**.
3. Select **Jira Issue Validator** from the list of plugins.
    * Enter a task name (mandatory).
    * Optionally, enter a description.
    * Provide values for the input variables.

    | Variable       | Format | Description                                               |
    | -------------- | ------ | --------------------------------------------------------- |
    | JiraUsername   | String | Your Jira username  (e.g., johndoe@devtron.ai)            |
    | JiraPassword   | String | Your Jira API token provided by the Jira admin            |
    | JiraBaseUrl    | String | The base URL of your Jira instance (e.g., https://yourdomain.atlassian.net) |

    * `Trigger/Skip Condition` allows you to set conditions under which this task will execute or be skipped.
    * `Pass/Failure Condition` allows you to define conditions that determine whether the build passes or fails based on Jira validation.

4. Go to the **Build Stage**.

5. Select **Pull Request** in the **Source Type** dropdown.

6. Use filters to fetch only the PRs matching your regex. Here are few examples:
    * **Title** can be a regex pattern (e.g., `^(?P<jira_Id>([a-zA-Z0-9-].*))`) to extract the Jira ID from the PR title. Only those PRs fulfilling the regex will be shown for image build process. 
    * **State** can be `^open$`, where only PRs in open state will be shown for image build process.

7. Click **Update Pipeline**.

--- 

## Results

**Case 1**: If Jira issue exists and the same is found in the PR title

![Figure 1: Jira Issue Match](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/plugins/jira/jira-issue-validator.jpg)

**Case 2**: If Jira issue is not found

![Figure 2: Error in Finding Jira Issue](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/plugins/jira/issue-validation-failed.jpg)
