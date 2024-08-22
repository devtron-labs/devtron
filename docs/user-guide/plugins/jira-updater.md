# Jira Issue Updater

## Introduction
The Jira Issue Updater plugin extends the capabilities of Devtron CI by allowing updates to Jira issues directly from the pipeline. It can add build pipeline status and docker image ID as a comment on Jira tickets, keeping the issue tracking synchronized with your CI processes.

### Prerequisites

- A Jira account with the necessary [API access](https://support.atlassian.com/atlassian-account/docs/manage-api-tokens-for-your-atlassian-account/#Create-an-API-token).
- The API credentials (username, password, and base URL) for your Jira instance. Obtain the API credentials from your Jira admin if required.
- A pull request raised with your Git provider. Title of pull request must contain the Jira ID.
- Jira Issue (e.g., REDOC-12)
- Webhook added to the git repository. [Click here](https://docs.devtron.ai/usage/applications/creating-application/workflow/ci-pipeline#configuring-webhook) to know more.

---

## Steps

1. On the **Edit build pipeline** page, go to the **Post-Build Stage**.
2. Click **+ Add task**.
3. Select **Jira Issue Updater** from the list of plugins.
    * Enter a task name (mandatory).
    * Optionally, enter a description.
    * Provide values for the input variables.

    | Variable                 | Format | Description                                               |
    | ------------------------ | ------ | --------------------------------------------------------- |
    | JiraUsername             | String | Your Jira username (e.g., johndoe@devtron.ai)             |
    | JiraPassword             | String | Your Jira API token provided by the Jira admin            |
    | JiraBaseUrl              | String | The base URL of your Jira instance (e.g., https://yourdomain.atlassian.net/) |
    | UpdateWithDockerImageId  | Bool   | Set to `true` to include the Docker Image ID in the update  |
    | UpdateWithBuildStatus    | Bool   | Set to `true` to include the build status in the update     |

    * `Trigger/Skip Condition` allows you to set conditions under which this task will execute or be skipped.
    * `Pass/Failure Condition` allows you define conditions to determine if the build passes or fails based on the Jira update.

4. Go to the **Build Stage**.

5. Select **Pull Request** in the **Source Type** dropdown.

6. Use filters to fetch only the PRs matching your regex. Here are few examples:
    * **Title** can be a regex pattern (e.g., `^(?P<jira_Id>([a-zA-Z0-9-].*))`) to extract the Jira ID from the PR title. Only those PRs fulfilling the regex will be shown for image build process. 
    * **State** can be `^open$`, where only PRs in open state will be shown for image build process.

7. Click **Update Pipeline**.

--- 

## Results

![Figure 1: Build Log](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/plugins/jira/jira-updater-log.jpg)

![Figure 2: Comments added by the Plugin on the Jira Issue](https://devtron-public-asset.s3.us-east-2.amazonaws.com/images/plugins/jira/jira-updater.jpg)





